package repository

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/test"
	"encoding/json"
	"fmt"
	eventstoredynamodb "github.com/j5ik2o/event-store-adapter-go/v5/dynamodb"
	"github.com/j5ik2o/event-store-adapter-go/v5/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
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

	var groupChatId *models.GroupChatId
	var adminId models.UserAccountId

	{
		adminId = models.NewUserAccountId()
		name := models.NewGroupChatName("test").MustGet()
		groupChat, event := domain.NewGroupChat(name, adminId)
		groupChatId = groupChat.GetGroupChatId()
		jsonObj, err := json.Marshal(event.ToJSON())
		require.NoError(t, err)
		slog.Info(fmt.Sprintf("event = %s", string(jsonObj)))

		err, b := repository.Store(event, &groupChat).Get()
		require.False(t, b)
	}

	{
		groupChat := repository.FindById(groupChatId).MustGet().MustGet()
		require.NotNil(t, groupChat)
		assert.Equal(t, groupChat.GetId(), groupChat.GetId())
		name2 := models.NewGroupChatName("test2").MustGet()
		renameResult := groupChat.Rename(name2, adminId).MustGet()
		groupChatUpdated := renameResult.V1
		event := renameResult.V2
		_, b := repository.Store(event, &groupChatUpdated).Get()
		require.False(t, b)
	}

	senderId := models.NewUserAccountId()

	{
		groupChat := repository.FindById(groupChatId).MustGet().MustGet()
		addMemberResult := groupChat.AddMember(models.NewMemberId(), senderId, models.MemberRole, adminId).MustGet()
		groupChatUpdated := addMemberResult.V1
		event := addMemberResult.V2
		_, b := repository.Store(event, &groupChatUpdated).Get()
		require.False(t, b)
	}

	{
		groupChat := repository.FindById(groupChatId).MustGet().MustGet()
		messageId := models.NewMessageId()
		message := models.NewMessage(messageId, "text", senderId).MustGet()
		postMessageResult := groupChat.PostMessage(message, senderId).MustGet()
		groupChatUpdated := postMessageResult.V1
		event := postMessageResult.V2
		_, b := repository.Store(event, &groupChatUpdated).Get()
		require.False(t, b)
	}
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
	jsonObj, err := json.Marshal(event.ToJSON())
	require.NoError(t, err)
	slog.Info(fmt.Sprintf("event = %s", string(jsonObj)))
	err, b := repository.StoreEventWithSnapshot(event, &groupChat).Get()
	require.NoError(t, err)
	require.False(t, b)

	groupChat2 := repository.FindById(groupChat.GetGroupChatId()).MustGet().MustGet()
	require.NotNil(t, groupChat2)
	assert.Equal(t, groupChat.GetId(), groupChat2.GetId())
}
