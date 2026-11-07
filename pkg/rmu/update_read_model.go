package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"encoding/json"
	"fmt"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	_ "github.com/go-sql-driver/mysql"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type ReadModelUpdater struct {
	dao GroupChatDao
}

// NewReadModelUpdater is a constructor for ReadModelUpdater.
func NewReadModelUpdater(dao GroupChatDao) ReadModelUpdater {
	return ReadModelUpdater{dao}
}

// UpdateReadModel processes events from DynamoDB stream and updates the read model.
func (r *ReadModelUpdater) UpdateReadModel(ctx context.Context, event dynamodbevents.DynamoDBEvent) error {
	for _, record := range event.Records {
		slog.Info("Processing request data for event GetId %s, type %s.", record.EventID, record.EventName)
		attributeValues := record.Change.NewImage
		payloadBytes := convertToBytes(attributeValues["payload"])
		typeValueStr, err := getTypeString(payloadBytes)
		if err != nil {
			return err
		}
		slog.Debug(fmt.Sprintf("typeValueStr = %s", typeValueStr))
		if !strings.HasPrefix(typeValueStr, "GroupChat") {
			continue
		}
		groupChatEvent, err := convertGroupChatEvent(payloadBytes)
		if err != nil {
			return err
		}
		groupChatId, err := getGroupChatId(attributeValues)
		if err != nil {
			return err
		}
		occurredAt, err := getOccurredAt(attributeValues)
		if err != nil {
			return err
		}

		switch ev := groupChatEvent.(type) {
		case *events.GroupChatCreated:
			err = createGroupChat(ev, groupChatId, occurredAt, r)
		case *events.GroupChatDeleted:
			err = deleteGroupChat(ev, groupChatId, occurredAt, r)
		case *events.GroupChatRenamed:
			err = renameGroupChat(ev, groupChatId, occurredAt, r)
		case *events.GroupChatMemberAdded:
			err = addMember(ev, groupChatId, occurredAt, r)
		case *events.GroupChatMemberRemoved:
			err = removeMember(ev, groupChatId, occurredAt, r)
		case *events.GroupChatMessagePosted:
			err = postMessage(ev, groupChatId, occurredAt, r)
		case *events.GroupChatMessageEdited:
			err = editMessage(ev, groupChatId, occurredAt, r)
		case *events.GroupChatMessageDeleted:
			err = deleteMessage(ev, groupChatId, occurredAt, r)
		}
		if err != nil {
			return err
		}

		// Print new values for attributes of type String
		for name, value := range record.Change.NewImage {
			slog.Debug(fmt.Sprintf("Attribute name: %s, value: %s", name, value.String()))
		}
	}
	return nil
}

func createGroupChat(ev *events.GroupChatCreated, groupChatId *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("createGroupChat: start: ev = %v", ev))
	name := ev.GetName()
	executorId := ev.GetExecutorId()
	err := r.dao.InsertGroupChat(groupChatId, name, executorId, occurredAt)
	if err != nil {
		return err
	}

	administrator := ev.GetMembers().GetAdministrator()
	memberId := administrator.GetId()
	userAccountId := administrator.GetUserAccountId()
	member := models.NewMember(*memberId, *userAccountId, models.AdminRole)
	err = r.dao.InsertMember(groupChatId, &member, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("createGroupChat: finished")
	return nil
}

func deleteGroupChat(_ *events.GroupChatDeleted, groupChatId *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info("deleteGroupChat: start")
	err := r.dao.DeleteGroupChat(groupChatId, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("deleteGroupChat: finished")
	return nil
}

func renameGroupChat(ev *events.GroupChatRenamed, groupChatId *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("renameGroupChat: start: ev = %v", ev))
	err := r.dao.UpdateName(groupChatId, ev.GetName(), occurredAt)
	if err != nil {
		return err
	}
	slog.Info("renameGroupChat: finished")
	return nil
}

func addMember(ev *events.GroupChatMemberAdded, groupChatId *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("addMember: start: ev = %v", ev))
	err := r.dao.InsertMember(groupChatId, ev.GetMember(), occurredAt)
	if err != nil {
		return err
	}
	slog.Info("addMember: finished")
	return nil
}

func removeMember(ev *events.GroupChatMemberRemoved, groupChatId *models.GroupChatId, _ time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("removeMember: start: ev = %v", ev))
	err := r.dao.DeleteMember(groupChatId, ev.GetUserAccountId())
	if err != nil {
		return err
	}
	slog.Info("removeMember: finished")
	return nil
}

func postMessage(ev *events.GroupChatMessagePosted, groupChatId *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("postMessage: start: ev = %v", ev))
	messageId := ev.GetMessage().GetId()
	userAccountId := ev.GetMessage().GetSenderId()
	text := ev.GetMessage().GetText()
	err := r.dao.InsertMessage(messageId, groupChatId, userAccountId, text, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("postMessage: finished")
	return nil
}

func editMessage(ev *events.GroupChatMessageEdited, _ *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("editMessage: start: ev = %v", ev))
	messageId := ev.GetMessage().GetId()
	text := ev.GetMessage().GetText()
	err := r.dao.UpdateMessage(messageId, text, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("editMessage: finished")
	return nil
}

func deleteMessage(ev *events.GroupChatMessageDeleted, _ *models.GroupChatId, occurredAt time.Time, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("deleteMessage: start: ev = %v", ev))
	messageId := ev.GetMessageId()
	err := r.dao.DeleteMessage(messageId, occurredAt)
	if err != nil {
		return err
	}
	slog.Info("deleteMessage: finished")
	return nil
}

func convertToBytes(payloadAttr dynamodbevents.DynamoDBAttributeValue) []byte {
	var payloadBytes []byte
	if payloadAttr.DataType() == dynamodbevents.DataTypeBinary {
		payloadBytes = payloadAttr.Binary()
	} else if payloadAttr.DataType() == dynamodbevents.DataTypeString {
		payloadBytes = []byte(payloadAttr.String())
	}
	return payloadBytes
}

func getTypeString(data []byte) (string, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		slog.Info(fmt.Sprintf("getTypeString: err = %v, %s", err, string(data)))
		return "", err
	}
	typeValue, ok := parsed["type"].(string)
	if !ok {
		return "", fmt.Errorf("event type is missing or not a string")
	}
	return typeValue, nil
}

func convertGroupChatEvent(payloadBytes []byte) (events.GroupChatEvent, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &parsed); err != nil {
		return nil, err
	}
	return repository.EventConverter(parsed)
}

func getGroupChatId(attributes map[string]dynamodbevents.DynamoDBAttributeValue) (*models.GroupChatId, error) {
	aidAttr, ok := attributes["aid"]
	if !ok || aidAttr.DataType() != dynamodbevents.DataTypeString {
		return nil, fmt.Errorf("aid is missing or not a string")
	}
	id, err := models.NewGroupChatIdFromString(aidAttr.String()).Get()
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func getOccurredAt(attributes map[string]dynamodbevents.DynamoDBAttributeValue) (time.Time, error) {
	occurredAtAttr, ok := attributes["occurred_at"]
	if !ok || occurredAtAttr.DataType() != dynamodbevents.DataTypeNumber {
		return time.Time{}, fmt.Errorf("occurred_at is missing or not a number")
	}
	nanos, err := strconv.ParseInt(occurredAtAttr.Number(), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, nanos), nil
}
