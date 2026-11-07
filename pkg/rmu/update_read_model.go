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

// ReadModelUpdater applies head stream records to the read model.
type ReadModelUpdater struct {
	applier  *Applier
	recovery *RecoveryService
	retry    RetryPolicy
}

type UpdaterOption func(*ReadModelUpdater)

func WithApplier(applier *Applier) UpdaterOption {
	return func(u *ReadModelUpdater) {
		u.applier = applier
	}
}

func WithRecoveryService(recovery *RecoveryService) UpdaterOption {
	return func(u *ReadModelUpdater) {
		u.recovery = recovery
	}
}

func WithRetryPolicy(policy RetryPolicy) UpdaterOption {
	return func(u *ReadModelUpdater) {
		u.retry = policy
	}
}

// NewReadModelUpdater is a constructor for ReadModelUpdater.
func NewReadModelUpdater(dao GroupChatDao, journal JournalReader, opts ...UpdaterOption) ReadModelUpdater {
	updater := ReadModelUpdater{
		applier: NewApplier(dao, journal, 1),
		retry:   NoRetryPolicy(),
	}
	for _, opt := range opts {
		opt(&updater)
	}
	return updater
}

// UpdateReadModel applies records in order and stops at the first failure.
// Lambda uses HandleDynamoDBEvent instead so the successful records in a batch
// are not replayed.
func (r *ReadModelUpdater) UpdateReadModel(ctx context.Context, event dynamodbevents.DynamoDBEvent) error {
	for _, record := range event.Records {
		if err := r.processRecord(ctx, record); err != nil {
			return fmt.Errorf("process head stream record %s: %w", record.EventID, err)
		}
	}
	return nil
}

// HandleDynamoDBEvent is the Lambda entry point. It returns the records that
// must be redelivered as a partial batch response. Records of an aggregate whose
// earlier event failed are also left for redelivery: the position rule forbids
// advancing past an unapplied event, so applying them would only fail again.
// The event source mapping must set FunctionResponseTypes to
// ReportBatchItemFailures and BisectBatchOnFunctionError to true.
func (r *ReadModelUpdater) HandleDynamoDBEvent(ctx context.Context, event dynamodbevents.DynamoDBEvent) (dynamodbevents.DynamoDBEventResponse, error) {
	response := dynamodbevents.DynamoDBEventResponse{}
	blocked := map[string]bool{}
	for _, record := range event.Records {
		aggregateID, ok := RecordAggregateID(record)
		if ok && blocked[aggregateID] {
			response.BatchItemFailures = append(response.BatchItemFailures, dynamodbevents.DynamoDBBatchItemFailure{
				ItemIdentifier: record.EventID,
			})
			continue
		}
		if err := r.processRecord(ctx, record); err != nil {
			response.BatchItemFailures = append(response.BatchItemFailures, dynamodbevents.DynamoDBBatchItemFailure{
				ItemIdentifier: record.EventID,
			})
			if ok {
				blocked[aggregateID] = true
			}
		}
	}
	return response, nil
}

func (r *ReadModelUpdater) processRecord(ctx context.Context, record dynamodbevents.DynamoDBEventRecord) error {
	identity, err := parseHeadRecord(record)
	if err != nil {
		if _, isPermanent := ClassifyFailure(err); isPermanent && r.recovery != nil && identity.aggregateID != "" {
			if recordErr := r.recovery.Record(ctx, identity.aggregateID, identity.seqNr, err); recordErr != nil {
				return &RecoveryRecordError{ApplyErr: err, RecordErr: recordErr}
			}
		}
		return err
	}
	if identity.skip {
		return nil
	}
	for _, event := range identity.events {
		applyErr := r.retry.Do(ctx, func() error {
			return r.applier.applyHeadEvent(ctx, identity.aggregateID, event)
		})
		if applyErr == nil {
			continue
		}
		if _, isPermanent := ClassifyFailure(applyErr); isPermanent && r.recovery != nil {
			if recordErr := r.recovery.Record(ctx, identity.aggregateID, event.seqNr, applyErr); recordErr != nil {
				return &RecoveryRecordError{ApplyErr: applyErr, RecordErr: recordErr}
			}
		}
		return applyErr
	}
	return nil
}

type recordIdentity struct {
	skip        bool
	aggregateID string
	seqNr       int64
	events      []headEvent
}

func parseHeadRecord(record dynamodbevents.DynamoDBEventRecord) (recordIdentity, error) {
	switch record.EventName {
	case "INSERT", "MODIFY":
	case "REMOVE":
		return recordIdentity{skip: true}, nil
	default:
		return recordIdentity{}, permanent(FailureKindEventPayload,
			fmt.Errorf("unexpected stream event name %q", record.EventName))
	}
	image := record.Change.NewImage
	if image == nil {
		return recordIdentity{}, permanent(FailureKindEventPayload, fmt.Errorf("head stream record has no new image"))
	}
	aggregateID, err := stringAttribute(image, "aid")
	if err != nil {
		return recordIdentity{}, permanent(FailureKindEventPayload, err)
	}
	if aggregateID == configAggregateID {
		return recordIdentity{skip: true}, nil
	}
	if !strings.HasPrefix(aggregateID, models.GroupChatIdPrefix+"-") {
		slog.Debug("skip head record of another aggregate type", "aid", aggregateID)
		return recordIdentity{skip: true}, nil
	}
	headSeqNr, _ := numberAttribute(image, "seq_nr")
	headEvents, err := parseHeadEvents(image)
	if err != nil {
		return recordIdentity{aggregateID: aggregateID, seqNr: headSeqNr},
			permanent(FailureKindEventPayload, err)
	}
	return recordIdentity{aggregateID: aggregateID, seqNr: headSeqNr, events: headEvents}, nil
}

func RecordAggregateID(record dynamodbevents.DynamoDBEventRecord) (string, bool) {
	if record.Change.NewImage == nil {
		return "", false
	}
	aggregateID, err := stringAttribute(record.Change.NewImage, "aid")
	if err != nil || aggregateID == configAggregateID ||
		!strings.HasPrefix(aggregateID, models.GroupChatIdPrefix+"-") {
		return "", false
	}
	return aggregateID, true
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
		return permanent(FailureKindEventPayload, err)
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
		return permanent(FailureKindUnsupportedEvent,
			fmt.Errorf("unsupported event type %s", groupChatEvent.GetTypeName()))
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
