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
	"strings"
	"time"
)

type ReadModelUpdater struct {
	dao GroupChatDao
}

type GroupChatDao interface {
	InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error
	DeleteGroupChat(id *models.GroupChatId, at time.Time) error
	UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, at time.Time) error

	InsertMember(id *models.MemberId, aggregateId *models.GroupChatId, accountId *models.UserAccountId, role models.Role, at time.Time) error
	DeleteMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId) error

	InsertMessage(id *models.MessageId, aggregateId *models.GroupChatId, userAccountId *models.UserAccountId, text string, at time.Time) error
	DeleteMessage(id *models.MessageId, at time.Time) error
}

// NewReadModelUpdater is a constructor for ReadModelUpdater.
func NewReadModelUpdater(dao GroupChatDao) *ReadModelUpdater {
	return &ReadModelUpdater{dao}
}

// UpdateReadModel processes events from DynamoDB stream and updates the read model.
func (r *ReadModelUpdater) UpdateReadModel(ctx context.Context, event dynamodbevents.DynamoDBEvent) error {
	for _, record := range event.Records {
		fmt.Printf("Processing request data for event GetId %s, type %s.\n", record.EventID, record.EventName)
		attributeValues := record.Change.NewImage
		payloadBytes := convertToBytes(attributeValues["payload"])
		typeValueStr, err := getTypeString(payloadBytes).Get()
		if err != nil {
			return err
		}
		fmt.Printf("typeValueStr = %s\n", typeValueStr)
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
			fmt.Printf("Attribute name: %s, value: %s\n", name, value.String())
		}
	}
	return nil
}

func createGroupChat(ev *events.GroupChatCreated, r *ReadModelUpdater) error {
	fmt.Printf("createGroupChat: start: ev = %v\n", ev)
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
	accountId := administrator.GetUserAccountId()
	err = r.dao.InsertMember(memberId, groupChatId, accountId, models.AdminRole, occurredAt)
	if err != nil {
		return err
	}
	fmt.Printf("createGroupChat: finished\n")
	return nil
}

func deleteGroupChat(ev *events.GroupChatDeleted, r *ReadModelUpdater) error {
	fmt.Printf("deleteGroupChat: start: ev = %v\n", ev)
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.DeleteGroupChat(groupChatId, occurredAt)
	if err != nil {
		return err
	}
	fmt.Printf("deleteGroupChat: finished\n")
	return nil
}

func renameGroupChat(ev *events.GroupChatRenamed, r *ReadModelUpdater) error {
	fmt.Printf("renameGroupChat: start: ev = %v\n", ev)
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	name := ev.GetName()
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.UpdateName(groupChatId, name, occurredAt)
	if err != nil {
		return err
	}
	fmt.Printf("renameGroupChat: finished\n")
	return nil
}

func addMember(ev *events.GroupChatMemberAdded, r *ReadModelUpdater) error {
	fmt.Printf("addMember: start: ev = %v\n", ev)
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	memberId := ev.GetMember().GetId()
	accountId := ev.GetMember().GetUserAccountId()
	role := ev.GetMember().GetRole()
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.InsertMember(memberId, groupChatId, accountId, role, occurredAt)
	if err != nil {
		return err
	}
	fmt.Printf("addMember: finished\n")
	return nil
}

func removeMember(ev *events.GroupChatMemberRemoved, r *ReadModelUpdater) error {
	fmt.Printf("removeMember: start: ev = %v\n", ev)
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	accountId := ev.GetUserAccountId()
	err := r.dao.DeleteMember(groupChatId, accountId)
	if err != nil {
		return err
	}
	fmt.Printf("removeMember: finished\n")
	return nil
}

func postMessage(ev *events.GroupChatMessagePosted, r *ReadModelUpdater) error {
	fmt.Printf("postMessage: start: ev = %v\n", ev)
	messageId := ev.GetMessage().GetId()
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	accountId := ev.GetMessage().GetSenderId()
	text := ev.GetMessage().GetText()
	createdAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.InsertMessage(messageId, groupChatId, accountId, text, createdAt)
	if err != nil {
		return err
	}
	fmt.Printf("postMessage: finished\n")
	return nil
}

func deleteMessage(ev *events.GroupChatMessageDeleted, r *ReadModelUpdater) error {
	fmt.Printf("deleteMessage: start: ev = %v\n", ev)
	messageId := ev.GetMessageId()
	updatedAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.DeleteMessage(messageId, updatedAt)
	if err != nil {
		return err
	}
	fmt.Printf("deleteMessage: finished\n")
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
		fmt.Printf("getTypeString: err = %v, %s\n", err, string(bytes))
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
