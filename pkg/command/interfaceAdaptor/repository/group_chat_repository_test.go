package repository

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/test"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	eventstoredynamodb "github.com/j5ik2o/event-store-adapter-go/v5/dynamodb"
	"github.com/j5ik2o/event-store-adapter-go/v5/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	journalTableName         = "journal"
	snapshotTableName        = "snapshot"
	headTableName            = "head"
	snapshotHistoryIndexName = "snapshot-history"
)

func Test_GroupChatRepository_OnDynamoDB_FindById(t *testing.T) {
	// Given
	ctx := context.Background()

	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)
	assert.NotNil(t, container)

	dynamodbClient, err := test.CreateDynamoDBClient(t, ctx, container)
	require.NoError(t, err)
	assert.NotNil(t, dynamodbClient)

	err = test.CreateJournalTable(t, ctx, dynamodbClient, journalTableName)
	require.NoError(t, err)

	err = test.CreateSnapshotTable(t, ctx, dynamodbClient, snapshotTableName, snapshotHistoryIndexName)
	require.NoError(t, err)

	err = test.CreateHeadTable(t, ctx, dynamodbClient, headTableName)
	require.NoError(t, err)

	eventStore, err := eventstoredynamodb.New(
		ctx,
		dynamodbClient,
		eventstoredynamodb.Config{
			JournalTableName:         journalTableName,
			SnapshotTableName:        snapshotTableName,
			HeadTableName:            headTableName,
			SnapshotHistoryIndexName: snapshotHistoryIndexName,
		},
		NewEventSerializer(),
		NewSnapshotSerializer())
	if err != nil {
		t.Fatal(err)
	}

	repository, err := NewGroupChatRepository(eventStore, WithRetention(2))
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()
	groupChat, createEvent := domain.NewGroupChat(name, adminId)
	groupChatId := groupChat.GetGroupChatId()
	staleSnapshot := groupChat

	_, b := repository.Store(createEvent, &groupChat).Get()
	require.False(t, b)

	{
		found := repository.FindById(groupChatId).MustGet().MustGet()
		name2 := models.NewGroupChatName("test2").MustGet()
		renameResult := found.Rename(name2, adminId).MustGet()
		_, b := repository.Store(renameResult.V2, &renameResult.V1).Get()
		require.False(t, b)
	}

	senderId := models.NewUserAccountId()

	{
		found := repository.FindById(groupChatId).MustGet().MustGet()
		addMemberResult := found.AddMember(models.NewMemberId(), senderId, models.MemberRole, adminId).MustGet()
		_, b := repository.Store(addMemberResult.V2, &addMemberResult.V1).Get()
		require.False(t, b)
	}

	{
		found := repository.FindById(groupChatId).MustGet().MustGet()
		messageId := models.NewMessageId()
		message := models.NewMessage(messageId, "text", senderId).MustGet()
		postMessageResult := found.PostMessage(message, senderId).MustGet()
		_, b := repository.Store(postMessageResult.V2, &postMessageResult.V1).Get()
		require.False(t, b)
	}

	replayed := repository.FindById(groupChatId).MustGet().MustGet()
	require.Equal(t, uint64(4), replayed.GetSeqNr())
	require.Equal(t, "test2", replayed.GetName().String())

	assertEnvelopeLayout(t, ctx, dynamodbClient, groupChatId, 4)

	// A write built from a stale aggregate must fail the head seq_nr condition.
	staleEvent := events.NewGroupChatRenamed(name, adminId)
	err, b = repository.StoreEvent(&staleEvent, &staleSnapshot).Get()
	require.True(t, b)
	require.Error(t, err)
}

func assertEnvelopeLayout(t *testing.T, ctx context.Context, client *dynamodb.Client, groupChatId *models.GroupChatId, expectedSeqNr int) {
	t.Helper()
	aid := groupChatId.AsString()

	journal, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(journalTableName),
		KeyConditionExpression: aws.String("aid = :aid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":aid": &types.AttributeValueMemberS{Value: aid},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, err)
	require.Len(t, journal.Items, expectedSeqNr)

	expectedManifests := []string{
		"GroupChatCreated@1",
		"GroupChatRenamed@1",
		"GroupChatMemberAdded@1",
		"GroupChatMessagePosted@1",
	}
	for i, item := range journal.Items {
		require.Equal(t, fmt.Sprintf("%d", i+1), item["seq_nr"].(*types.AttributeValueMemberN).Value)
		require.Equal(t, expectedManifests[i], item["manifest"].(*types.AttributeValueMemberS).Value)
		require.NotEmpty(t, item["occurred_at"].(*types.AttributeValueMemberN).Value)

		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(item["payload"].(*types.AttributeValueMemberB).Value, &payload))
		require.Contains(t, payload, "type")
		for _, key := range []string{"type_name", "aggregate_id", "seq_nr", "occurred_at"} {
			require.NotContains(t, payload, key)
		}
	}

	head, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(headTableName),
		Key: map[string]types.AttributeValue{
			"aid": &types.AttributeValueMemberS{Value: aid},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, err)
	require.NotNil(t, head.Item)
	require.Equal(t, fmt.Sprintf("%d", expectedSeqNr), head.Item["seq_nr"].(*types.AttributeValueMemberN).Value)
	headEvents := head.Item["events"].(*types.AttributeValueMemberL).Value
	require.Len(t, headEvents, 1)
	headEvent := headEvents[0].(*types.AttributeValueMemberM).Value
	require.Equal(t, fmt.Sprintf("%d", expectedSeqNr), headEvent["seq_nr"].(*types.AttributeValueMemberN).Value)
	require.Equal(t, "GroupChatMessagePosted@1", headEvent["manifest"].(*types.AttributeValueMemberS).Value)

	snapshot, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(snapshotTableName),
		Key: map[string]types.AttributeValue{
			"aid":  &types.AttributeValueMemberS{Value: aid},
			"skey": &types.AttributeValueMemberN{Value: "0"},
		},
		ConsistentRead: aws.Bool(true),
	})
	require.NoError(t, err)
	require.NotNil(t, snapshot.Item)
	require.Equal(t, SnapshotManifest, snapshot.Item["manifest"].(*types.AttributeValueMemberS).Value)
	var snapshotPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(snapshot.Item["payload"].(*types.AttributeValueMemberB).Value, &snapshotPayload))
	require.NotContains(t, snapshotPayload, "version")
	require.Contains(t, snapshotPayload, "seq_nr")
}

func Test_GroupChatRepository_OnMemory_FindById(t *testing.T) {
	store, err := memory.NewStore()
	require.NoError(t, err)
	eventStore, err := memory.New(store, NewEventSerializer(), NewSnapshotSerializer())
	require.NoError(t, err)

	repository, err := NewGroupChatRepository(eventStore)
	require.NoError(t, err)
	adminId := models.NewUserAccountId()
	name, err := models.NewGroupChatName("test").Get()
	if err != nil {
		t.Fatal(err)
	}
	groupChat, event := domain.NewGroupChat(name, adminId)
	err, b := repository.StoreEventWithSnapshot(event, &groupChat).Get()
	require.NoError(t, err)
	require.False(t, b)

	groupChat2 := repository.FindById(groupChat.GetGroupChatId()).MustGet().MustGet()
	require.NotNil(t, groupChat2)
	assert.Equal(t, groupChat.GetId(), groupChat2.GetId())
}
