package repository

import (
	"context"
	"cqrs-es-example-go/domain"
	"cqrs-es-example-go/domain/events"
	"cqrs-es-example-go/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/j5ik2o/event-store-adapter-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"testing"
)

func TestGroupChatRepositoryImpl_FindById(t *testing.T) {
	// Given
	ctx := context.Background()
	container, err := localstack.RunContainer(
		ctx,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:2.1.0",
				Env: map[string]string{
					"SERVICES":              "dynamodb",
					"DEFAULT_REGION":        "us-east-1",
					"EAGER_SERVICE_LOADING": "1",
					"DYNAMODB_SHARED_DB":    "1",
					"DYNAMODB_IN_MEMORY":    "1",
				},
			},
		}),
	)
	require.Nil(t, err)
	assert.NotNil(t, container)
	dynamodbClient, err := common.CreateDynamoDBClient(t, ctx, container)
	require.Nil(t, err)
	assert.NotNil(t, dynamodbClient)
	journalTableName := "journal"
	journalAidIndexName := "journal-aid-index"
	err = common.CreateJournalTable(t, ctx, dynamodbClient, journalTableName, journalAidIndexName)
	require.Nil(t, err)
	snapshotTableName := "snapshot"
	snapshotAidIndexName := "snapshot-aid-index"
	err = common.CreateSnapshotTable(t, ctx, dynamodbClient, snapshotTableName, snapshotAidIndexName)
	require.Nil(t, err)

	// time.Sleep(5 * time.Second)

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
		journalTableName, snapshotTableName, journalAidIndexName, snapshotAidIndexName,
		32,
		eventConverter, aggregateConverter,
		esa.WithEventSerializer(&EventSerializer{}),
		esa.WithSnapshotSerializer(&SnapshotSerializer{}))

	if err != nil {
		t.Fatal(err)
	}
	repository := NewGroupChatRepository(eventStore)
	adminId := models.NewUserAccountId()
	name, err := models.NewGroupChatName("test")
	if err != nil {
		t.Fatal(err)
	}
	groupChat, event := domain.NewGroupChat(name, adminId, adminId)
	err = repository.StoreEventWithSnapshot(event, groupChat)
	require.NoError(t, err)

	groupChat2 := repository.FindById(groupChat.GetGroupChatId()).MustGet()
	require.NotNil(t, groupChat2)
	assert.Equal(t, groupChat.GetId(), groupChat2.GetId())
}
