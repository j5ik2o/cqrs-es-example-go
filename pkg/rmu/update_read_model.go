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
	Create(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error
	AddMember(id *models.MemberId, aggregateId *models.GroupChatId, accountId *models.UserAccountId, role models.Role, at time.Time) error
}

func NewReadModelUpdater(dao GroupChatDao) *ReadModelUpdater {
	return &ReadModelUpdater{dao}
}

func (r *ReadModelUpdater) UpdateReadModel(ctx context.Context, event dynamodbevents.DynamoDBEvent) error {
	for _, record := range event.Records {
		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
		attributeValues := record.Change.NewImage
		payloadBytes := convertToBytes(attributeValues["payload"])
		typeValueStr, err := getTypeString(payloadBytes).Get()
		if err != nil {
			return err
		}
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
			case *events.GroupChatRenamed:
			case *events.GroupChatMemberAdded:
			case *events.GroupChatMemberRemoved:
			case *events.GroupChatMessagePosted:
			case *events.GroupChatMessageDeleted:
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
	groupChatId := ev.GetAggregateId().(*models.GroupChatId)
	name := ev.GetName()
	executorId := ev.GetExecutorId()
	occurredAt := convertToTime(ev.GetOccurredAt())
	err := r.dao.Create(groupChatId, name, executorId, occurredAt)
	if err != nil {
		return err
	}
	administrator := ev.GetMembers().GetAdministrator()
	memberId := administrator.GetId()
	accountId := administrator.GetUserAccountId()
	err = r.dao.AddMember(memberId, groupChatId, accountId, models.AdminRole, occurredAt)
	if err != nil {
		return err
	}
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
