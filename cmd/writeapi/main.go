package main

import (
	"context"
	"cqrs-es-example-go/domain"
	"cqrs-es-example-go/domain/events"
	"cqrs-es-example-go/domain/models"
	controller "cqrs-es-example-go/interfaceAdapter/ctrl"
	"cqrs-es-example-go/repository"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"k8s.io/utils/env"
)

func main() {
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	dynamodbClient := dynamodb.NewFromConfig(awsCfg)
	journalTableName := env.GetString("JOURNAL_TABLE_NAME", "journal")
	snapshotTableName := env.GetString("SNAPSHOT_TABLE_NAME", "snapshot")
	journalAidIndexName := env.GetString("JOURNAL_AID_INDEX_NAME", "journal-aid-index")
	snapshotAidIndexName := env.GetString("SNAPSHOT_AID_INDEX_NAME", "snapshot-aid-index")
	shardCount, err := env.GetInt("SHARD_COUNT", 10)
	if err != nil {
		panic(err)
	}

	eventConverter := func(m map[string]interface{}) (esa.Event, error) {
		eventId := m["Id"].(string)
		groupChatId := models.ConvertGroupChatIdFromJSON(m["AggregateId"].(map[string]interface{}))
		groupChatName, err := models.ConvertGroupChatNameFromJSON(m["Name"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		members := models.ConvertMembersFromJSON(m["Members"].(map[string]interface{}))
		executorId, err := models.ConvertUserAccountIdFromJSON(m["ExecutorId"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		seqNr := uint64(m["SeqNr"].(float64))
		occurredAt := uint64(m["OccurredAt"].(float64))
		switch m["TypeName"].(string) {
		case "GroupChatCreated":
			return events.NewGroupChatCreatedFrom(
				eventId,
				groupChatId,
				groupChatName,
				members,
				seqNr,
				executorId,
				occurredAt,
			), nil
		case "GroupChatDeleted":
			return events.NewGroupChatDeletedFrom(
				eventId,
				groupChatId,
				seqNr,
				executorId,
				occurredAt,
			), nil
		case "GroupChatRenamed":
			name, err := models.NewGroupChatName(m["Name"].(string))
			if err != nil {
				return nil, err
			}
			return events.NewGroupChatRenamedFrom(
				eventId,
				groupChatId,
				name,
				seqNr,
				executorId,
				occurredAt,
			), nil
		case "GroupChatMemberAdded":
			memberObj := m["Member"].(map[string]interface{})
			memberId := models.ConvertMemberIdFromJSON(memberObj["MemberId"].(map[string]interface{}))
			userAccountId, err := models.ConvertUserAccountIdFromJSON(memberObj["UserAccountId"].(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			role := models.Role(memberObj["Role"].(int))
			member := models.NewMember(memberId, userAccountId, role)
			return events.NewGroupChatMemberAddedFrom(
				eventId,
				groupChatId,
				member,
				seqNr,
				executorId,
				occurredAt,
			), nil
		case "GroupChatMemberRemoved":
			userAccountId, err := models.ConvertUserAccountIdFromJSON(m["UserAccountId"].(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			return events.NewGroupChatMemberRemovedFrom(
				eventId,
				groupChatId,
				userAccountId,
				seqNr,
				executorId,
				occurredAt,
			), nil
		case "GroupChatMessagePosted":
			message := models.ConvertMessageFromJSON(m["Message"].(map[string]interface{}))
			return events.NewGroupChatMessagePostedFrom(
				eventId,
				groupChatId,
				message,
				seqNr,
				executorId,
				occurredAt,
			), nil
		case "GroupChatMessageDeleted":
			messageId := models.ConvertMessageIdFromJSON(m["MessageId"].(map[string]interface{}))
			return events.NewGroupChatMessageDeletedFrom(
				eventId,
				groupChatId,
				messageId,
				seqNr,
				executorId,
				occurredAt,
			), nil
		default:
			return nil, fmt.Errorf("unknown event type")
		}
	}

	aggregateConverter := func(m map[string]interface{}) (esa.Aggregate, error) {
		groupChatId := models.ConvertGroupChatIdFromJSON(m["Id"].(map[string]interface{}))
		name, err := models.ConvertGroupChatNameFromJSON(m["Name"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		members := models.ConvertMembersFromJSON(m["Members"].(map[string]interface{}))
		messages := models.ConvertMessagesFromJSON(m["Messages"].(map[string]interface{}))
		seqNr := uint64(m["SeqNr"].(float64))
		version := uint64(m["Version"].(float64))
		deleted := m["Deleted"].(bool)
		result := domain.NewGroupChatFrom(groupChatId, name, members, messages, seqNr, version, deleted)
		return result, nil
	}

	eventStore, err := esa.NewEventStore(
		dynamodbClient,
		journalTableName,
		snapshotTableName,
		journalAidIndexName, snapshotAidIndexName,
		uint64(shardCount),
		eventConverter,
		aggregateConverter,
		esa.WithEventSerializer(repository.NewEventSerializer()),
		esa.WithSnapshotSerializer(repository.NewSnapshotSerializer()))

	groupChatRepository := repository.NewGroupChatRepository(eventStore)
	groupChatController := controller.NewGroupChatController(groupChatRepository)

	engine := gin.Default()
	engine.GET("/", controller.Index)
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
	}
	err = engine.Run()
	if err != nil {
		panic(err)
	}
}
