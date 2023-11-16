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
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/samber/mo"
	"strings"
	"time"
)

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

func UpdateReadModel(ctx context.Context, event dynamodbevents.DynamoDBEvent) {
	dbUrl := env.String("", "DATABASE_URL")
	dataSourceName := fmt.Sprintf("%s?parseTime=true", dbUrl)
	db, err := sqlx.Connect("mysql", dataSourceName)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)
	if err != nil {
		panic(err.Error())
	}
	dao := NewGroupChatDao(db)
	for _, record := range event.Records {
		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
		attributeValues := record.Change.NewImage
		payloadAttr := attributeValues["payload"]
		payloadBytes := []byte(payloadAttr.Binary())
		typeValueStr, err := getTypeString(payloadBytes).Get()
		if err != nil {
			panic(err.Error())
		}
		if strings.HasPrefix(typeValueStr, "GroupChat") {
			event, err := convertGroupChatEvent(payloadBytes).Get()
			if err != nil {
				panic(err.Error())
			}
			switch event.(type) {
			case *events.GroupChatCreated:
				ev := event.(*events.GroupChatCreated)
				groupChatId := ev.GetAggregateId().(*models.GroupChatId)
				name := ev.GetName()
				executorId := ev.GetExecutorId()
				occurredAtUnix := int64(ev.GetOccurredAt()) * int64(time.Millisecond)
				occurredAt := time.Unix(0, occurredAtUnix)
				fmt.Printf("occurredAt = %v\n", occurredAt)
				err := dao.Create(groupChatId, name, executorId, occurredAt)
				if err != nil {
					panic(err.Error())
				}
				memberId := ev.GetMembers().GetAdministrator().GetId()
				accountId := ev.GetMembers().GetAdministrator().GetUserAccountId()
				err = dao.AddMember(memberId, groupChatId, accountId, models.AdminRole, occurredAt)
				if err != nil {
					panic(err.Error())
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
}
