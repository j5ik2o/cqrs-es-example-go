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
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/samber/mo"
	"log/slog"
	"strings"
	"time"
)

type ReadModelUpdater struct {
	dao GroupChatDao
}

type GroupChatDao interface {
	InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error
	DeleteGroupChat(aggregateId *models.GroupChatId, updatedAt time.Time) error
	UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, updatedAt time.Time) error

	InsertMember(aggregateId *models.GroupChatId, member *models.Member, createdAt time.Time) error
	DeleteMember(aggregateId *models.GroupChatId, userAccountId *models.UserAccountId) error

	InsertMessage(messageId *models.MessageId, aggregateId *models.GroupChatId, userAccountId *models.UserAccountId, text string, createdAt time.Time) error
	DeleteMessage(messageId *models.MessageId, updatedAt time.Time) error
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
		typeValueStr, err := getTypeString(payloadBytes).Get()
		if err != nil {
			return err
		}
		slog.Debug(fmt.Sprintf("typeValueStr = %s", typeValueStr))
		if strings.HasPrefix(typeValueStr, "GroupChat") {
			event, err := convertGroupChatEvent(payloadBytes).Get()
			if err != nil {
				return err
			}
			switch event.(type) {
			case *events.GroupChatCreated:
				ev := event.(*events.GroupChatCreated)
				err2 := createGroupChat(ev, r)
				if err2 != nil {
					return err2
				}
			case *events.GroupChatDeleted:
				ev := event.(*events.GroupChatDeleted)
				err2 := deleteGroupChat(ev, r)
				if err2 != nil {
					return err2
				}
			case *events.GroupChatRenamed:
				ev := event.(*events.GroupChatRenamed)
				err2 := renameGroupChat(ev, r)
				if err2 != nil {
					return err2
				}
			case *events.GroupChatMemberAdded:
				ev := event.(*events.GroupChatMemberAdded)
				err2 := addMember(ev, r)
				if err2 != nil {
					return err2
				}
			case *events.GroupChatMemberRemoved:
				ev := event.(*events.GroupChatMemberRemoved)
				err2 := removeMember(ev, r)
				if err2 != nil {
					return err2
				}
			case *events.GroupChatMessagePosted:
				ev := event.(*events.GroupChatMessagePosted)
				err2 := postMessage(ev, r)
				if err2 != nil {
					return err2
				}
			case *events.GroupChatMessageDeleted:
				ev := event.(*events.GroupChatMessageDeleted)
				err2 := deleteMessage(ev, r)
				if err2 != nil {
					return err2
				}
			default:
			}
		}
		// Print new values for attributes of type String
		for name, value := range record.Change.NewImage {
			slog.Debug(fmt.Sprintf("Attribute name: %s, value: %s", name, value.String()))
		}
	}
	return nil
}

func createGroupChat(ev *events.GroupChatCreated, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("createGroupChat: start: ev = %v", ev))
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	name := ev.GetName()
	executorId := ev.GetExecutorId()
	occurredAt := convertToTime(ev.GetOccurredAt())
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

func deleteGroupChat(ev *events.GroupChatDeleted, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("deleteGroupChat: start: ev = %v", ev))
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.DeleteGroupChat(groupChatId, occurredAt)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("deleteGroupChat: finished"))
	return nil
}

func renameGroupChat(ev *events.GroupChatRenamed, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("renameGroupChat: start: ev = %v", ev))
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	name := ev.GetName()
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.UpdateName(groupChatId, name, occurredAt)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("renameGroupChat: finished"))
	return nil
}

func addMember(ev *events.GroupChatMemberAdded, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("addMember: start: ev = %v", ev))
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.InsertMember(groupChatId, ev.GetMember(), occurredAt)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("addMember: finished"))
	return nil
}

func removeMember(ev *events.GroupChatMemberRemoved, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("removeMember: start: ev = %v", ev))
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	userAccountId := ev.GetUserAccountId()
	err := r.dao.DeleteMember(groupChatId, userAccountId)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("removeMember: finished"))
	return nil
}

func postMessage(ev *events.GroupChatMessagePosted, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("postMessage: start: ev = %v", ev))
	messageId := ev.GetMessage().GetId()
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	userAccountId := ev.GetMessage().GetSenderId()
	text := ev.GetMessage().GetText()
	createdAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.InsertMessage(messageId, groupChatId, userAccountId, text, createdAt)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("postMessage: finished"))
	return nil
}

func deleteMessage(ev *events.GroupChatMessageDeleted, r *ReadModelUpdater) error {
	slog.Info(fmt.Sprintf("deleteMessage: start: ev = %v", ev))
	messageId := ev.GetMessageId()
	updatedAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.DeleteMessage(messageId, updatedAt)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("deleteMessage: finished"))
	return nil
}

func convertToTime(epoc uint64) time.Time {
	occurredAtUnix := int64(epoc) * int64(time.Millisecond)
	occurredAt := time.Unix(0, occurredAtUnix)
	return occurredAt
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

func getTypeString(bytes []byte) mo.Result[string] {
	var parsed map[string]interface{}
	err := json.Unmarshal(bytes, &parsed)
	if err != nil {
		slog.Info(fmt.Sprintf("getTypeString: err = %v, %s", err, string(bytes)))
		return mo.Err[string](err)
	}
	typeValue, ok := parsed["type_name"].(string)
	if !ok {
		mo.Err[string](fmt.Errorf("type is not a string"))
	}
	return mo.Ok(typeValue)
}

func convertGroupChatEvent(payloadBytes []byte) mo.Result[esa.Event] {
	var parsed map[string]interface{}
	err := json.Unmarshal(payloadBytes, &parsed)
	if err != nil {
		mo.Err[esa.Event](err)
	}
	event, err := repository.EventConverter(parsed)
	if err != nil {
		mo.Err[esa.Event](err)
	}
	return mo.Ok(event)
}
