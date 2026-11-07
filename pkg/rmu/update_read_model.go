package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	dynamodbevents "github.com/aws/aws-lambda-go/events"
)

// configAggregateID is the reserved head item that carries the store
// configuration. It is not an aggregate and must be skipped.
const configAggregateID = "__config__"

type ReadModelUpdater struct {
	dao     GroupChatDao
	journal JournalReader
}

// NewReadModelUpdater is a constructor for ReadModelUpdater.
func NewReadModelUpdater(dao GroupChatDao, journal JournalReader) ReadModelUpdater {
	return ReadModelUpdater{dao: dao, journal: journal}
}

// UpdateReadModel applies the head stream records to the read model. Each record
// carries the head item's new image, whose events list holds the events written
// by one transaction. INSERT and MODIFY both carry events; a REMOVE never
// happens because the store does not delete heads.
func (r *ReadModelUpdater) UpdateReadModel(ctx context.Context, event dynamodbevents.DynamoDBEvent) error {
	for _, record := range event.Records {
		if err := r.applyRecord(ctx, record); err != nil {
			return fmt.Errorf("process head stream record %s: %w", record.EventID, err)
		}
	}
	return nil
}

func (r *ReadModelUpdater) applyRecord(ctx context.Context, record dynamodbevents.DynamoDBEventRecord) error {
	switch record.EventName {
	case "INSERT", "MODIFY":
	case "REMOVE":
		return nil
	default:
		return fmt.Errorf("unexpected stream event name %q", record.EventName)
	}
	image := record.Change.NewImage
	if image == nil {
		return fmt.Errorf("head stream record has no new image")
	}
	aggregateID, err := stringAttribute(image, "aid")
	if err != nil {
		return err
	}
	if aggregateID == configAggregateID {
		return nil
	}
	if !strings.HasPrefix(aggregateID, models.GroupChatIdPrefix+"-") {
		slog.Debug("skip head record of another aggregate type", "aid", aggregateID)
		return nil
	}
	headEvents, err := parseHeadEvents(image)
	if err != nil {
		return err
	}
	for _, headEvent := range headEvents {
		if err := r.applyHeadEvent(ctx, aggregateID, headEvent); err != nil {
			return err
		}
	}
	return nil
}

// headEvent is one envelope entry in the head item's events list.
type headEvent struct {
	seqNr      int64
	occurredAt time.Time
	manifest   string
	payload    []byte
}

// applyHeadEvent applies the stream event and, when the aggregate's position is
// behind, the journal events that fill the gap. Duplicate and stale sequence
// numbers are dropped. The read model changes and the position advance commit
// in one transaction.
func (r *ReadModelUpdater) applyHeadEvent(ctx context.Context, aggregateID string, event headEvent) error {
	groupChatID, err := models.NewGroupChatIdFromString(aggregateID).Get()
	if err != nil {
		return err
	}

	currentSeqNr, err := r.dao.GetPosition(ctx, aggregateID)
	if err != nil {
		return err
	}
	if event.seqNr <= currentSeqNr {
		slog.Debug("skip duplicate or stale event", "aid", aggregateID, "seq_nr", event.seqNr, "position", currentSeqNr)
		return nil
	}

	pending := []headEvent{event}
	if event.seqNr > currentSeqNr+1 {
		if r.journal == nil {
			return fmt.Errorf("aggregate %s: seq_nr %d follows position %d and no journal reader is configured", aggregateID, event.seqNr, currentSeqNr)
		}
		journalEvents, err := r.journal.ReadEvents(ctx, aggregateID, currentSeqNr+1, event.seqNr-1)
		if err != nil {
			return fmt.Errorf("aggregate %s: read journal seq_nr %d..%d: %w", aggregateID, currentSeqNr+1, event.seqNr-1, err)
		}
		if err := requireContiguous(currentSeqNr+1, event.seqNr-1, journalEvents); err != nil {
			return fmt.Errorf("aggregate %s: %w", aggregateID, err)
		}
		gap := make([]headEvent, len(journalEvents))
		for i, journalEvent := range journalEvents {
			gap[i] = headEvent{
				seqNr:      journalEvent.SeqNr,
				occurredAt: journalEvent.OccurredAt,
				manifest:   journalEvent.Manifest,
				payload:    journalEvent.Payload,
			}
		}
		pending = append(gap, event)
	}

	return r.dao.Process(ctx, aggregateID, func(w ReadModelWriter, position int64) (int64, error) {
		for _, pendingEvent := range pending {
			if pendingEvent.seqNr <= position {
				continue
			}
			if pendingEvent.seqNr != position+1 {
				return position, fmt.Errorf("expected seq_nr %d but got %d", position+1, pendingEvent.seqNr)
			}
			if err := applyGroupChatEvent(&groupChatID, pendingEvent, w); err != nil {
				return position, fmt.Errorf("apply seq_nr %d: %w", pendingEvent.seqNr, err)
			}
			position = pendingEvent.seqNr
		}
		return position, nil
	})
}

func requireContiguous(fromSeqNr, toSeqNr int64, events []JournalEvent) error {
	expected := fromSeqNr
	for _, event := range events {
		if event.SeqNr != expected {
			return fmt.Errorf("journal gap: expected seq_nr %d but got %d", expected, event.SeqNr)
		}
		expected++
	}
	if expected != toSeqNr+1 {
		return fmt.Errorf("journal is missing seq_nr %d..%d", expected, toSeqNr)
	}
	return nil
}

func parseHeadEvents(image map[string]dynamodbevents.DynamoDBAttributeValue) ([]headEvent, error) {
	rawEvents, ok := image["events"]
	if !ok {
		return nil, fmt.Errorf("head image has no events attribute")
	}
	if rawEvents.DataType() != dynamodbevents.DataTypeList {
		return nil, fmt.Errorf("head events attribute is not a list")
	}
	list := rawEvents.List()
	result := make([]headEvent, 0, len(list))
	for i, item := range list {
		if item.DataType() != dynamodbevents.DataTypeMap {
			return nil, fmt.Errorf("head event at index %d is not a map", i)
		}
		fields := item.Map()
		seqNr, err := numberAttribute(fields, "seq_nr")
		if err != nil {
			return nil, fmt.Errorf("head event at index %d: %w", i, err)
		}
		occurredAt, err := numberAttribute(fields, "occurred_at")
		if err != nil {
			return nil, fmt.Errorf("head event at index %d: %w", i, err)
		}
		manifest, err := stringAttribute(fields, "manifest")
		if err != nil {
			return nil, fmt.Errorf("head event at index %d: %w", i, err)
		}
		payload, err := bytesAttribute(fields, "payload")
		if err != nil {
			return nil, fmt.Errorf("head event at index %d: %w", i, err)
		}
		result = append(result, headEvent{
			seqNr:      seqNr,
			occurredAt: time.Unix(0, occurredAt),
			manifest:   manifest,
			payload:    payload,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].seqNr < result[j].seqNr })
	return result, nil
}

func stringAttribute(attributes map[string]dynamodbevents.DynamoDBAttributeValue, name string) (string, error) {
	attribute, ok := attributes[name]
	if !ok {
		return "", fmt.Errorf("attribute %q is missing", name)
	}
	if attribute.DataType() != dynamodbevents.DataTypeString {
		return "", fmt.Errorf("attribute %q is not a string", name)
	}
	return attribute.String(), nil
}

func numberAttribute(attributes map[string]dynamodbevents.DynamoDBAttributeValue, name string) (int64, error) {
	attribute, ok := attributes[name]
	if !ok {
		return 0, fmt.Errorf("attribute %q is missing", name)
	}
	if attribute.DataType() != dynamodbevents.DataTypeNumber {
		return 0, fmt.Errorf("attribute %q is not a number", name)
	}
	return strconv.ParseInt(attribute.Number(), 10, 64)
}

func bytesAttribute(attributes map[string]dynamodbevents.DynamoDBAttributeValue, name string) ([]byte, error) {
	attribute, ok := attributes[name]
	if !ok {
		return nil, fmt.Errorf("attribute %q is missing", name)
	}
	switch attribute.DataType() {
	case dynamodbevents.DataTypeBinary:
		return attribute.Binary(), nil
	case dynamodbevents.DataTypeString:
		return []byte(attribute.String()), nil
	default:
		return nil, fmt.Errorf("attribute %q is neither binary nor string", name)
	}
}

func applyGroupChatEvent(groupChatId *models.GroupChatId, event headEvent, w ReadModelWriter) error {
	groupChatEvent, err := convertGroupChatEvent(event.payload)
	if err != nil {
		return err
	}
	switch ev := groupChatEvent.(type) {
	case *events.GroupChatCreated:
		return createGroupChat(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatDeleted:
		return deleteGroupChat(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatRenamed:
		return renameGroupChat(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatMemberAdded:
		return addMember(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatMemberRemoved:
		return removeMember(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatMessagePosted:
		return postMessage(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatMessageEdited:
		return editMessage(ev, groupChatId, event.occurredAt, w)
	case *events.GroupChatMessageDeleted:
		return deleteMessage(ev, groupChatId, event.occurredAt, w)
	default:
		return fmt.Errorf("unsupported event type %s", groupChatEvent.GetTypeName())
	}
}

func createGroupChat(ev *events.GroupChatCreated, groupChatId *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("createGroupChat: start: ev = %v", ev))
	name := ev.GetName()
	executorId := ev.GetExecutorId()
	err := w.InsertGroupChat(groupChatId, name, executorId, occurredAt)
	if err != nil {
		return err
	}

	administrator := ev.GetMembers().GetAdministrator()
	memberId := administrator.GetId()
	userAccountId := administrator.GetUserAccountId()
	member := models.NewMember(*memberId, *userAccountId, models.AdminRole)
	err = w.InsertMember(groupChatId, &member, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("createGroupChat: finished")
	return nil
}

func deleteGroupChat(_ *events.GroupChatDeleted, groupChatId *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info("deleteGroupChat: start")
	err := w.DeleteGroupChat(groupChatId, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("deleteGroupChat: finished")
	return nil
}

func renameGroupChat(ev *events.GroupChatRenamed, groupChatId *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("renameGroupChat: start: ev = %v", ev))
	err := w.UpdateName(groupChatId, ev.GetName(), occurredAt)
	if err != nil {
		return err
	}
	slog.Info("renameGroupChat: finished")
	return nil
}

func addMember(ev *events.GroupChatMemberAdded, groupChatId *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("addMember: start: ev = %v", ev))
	err := w.InsertMember(groupChatId, ev.GetMember(), occurredAt)
	if err != nil {
		return err
	}
	slog.Info("addMember: finished")
	return nil
}

func removeMember(ev *events.GroupChatMemberRemoved, groupChatId *models.GroupChatId, _ time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("removeMember: start: ev = %v", ev))
	err := w.DeleteMember(groupChatId, ev.GetUserAccountId())
	if err != nil {
		return err
	}
	slog.Info("removeMember: finished")
	return nil
}

func postMessage(ev *events.GroupChatMessagePosted, groupChatId *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("postMessage: start: ev = %v", ev))
	messageId := ev.GetMessage().GetId()
	userAccountId := ev.GetMessage().GetSenderId()
	text := ev.GetMessage().GetText()
	err := w.InsertMessage(messageId, groupChatId, userAccountId, text, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("postMessage: finished")
	return nil
}

func editMessage(ev *events.GroupChatMessageEdited, _ *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("editMessage: start: ev = %v", ev))
	messageId := ev.GetMessage().GetId()
	text := ev.GetMessage().GetText()
	err := w.UpdateMessage(messageId, text, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("editMessage: finished")
	return nil
}

func deleteMessage(ev *events.GroupChatMessageDeleted, _ *models.GroupChatId, occurredAt time.Time, w ReadModelWriter) error {
	slog.Info(fmt.Sprintf("deleteMessage: start: ev = %v", ev))
	messageId := ev.GetMessageId()
	err := w.DeleteMessage(messageId, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("deleteMessage: finished")
	return nil
}

func convertGroupChatEvent(payloadBytes []byte) (events.GroupChatEvent, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &parsed); err != nil {
		return nil, err
	}
	return repository.EventConverter(parsed)
}
