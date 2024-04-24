package repository

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/test"
	"encoding/json"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
	"github.com/j5ik2o/event-store-adapter-go/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

var (
	journalTableName     = "journal"
	journalAidIndexName  = "journal-aid-index"
	snapshotTableName    = "snapshot"
	snapshotAidIndexName = "snapshot-aid-index"
)

func Test_GroupChatRepository_OnDynamoDB_FindById(t *testing.T) {
	// Given
	ctx := context.Background()

	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)
	assert.NotNil(t, container)

	dynamodbClient, err := common.CreateDynamoDBClient(t, ctx, container)
	require.NoError(t, err)
	assert.NotNil(t, dynamodbClient)

	err = common.CreateJournalTable(t, ctx, dynamodbClient, journalTableName, journalAidIndexName)
	require.NoError(t, err)

	err = common.CreateSnapshotTable(t, ctx, dynamodbClient, snapshotTableName, snapshotAidIndexName)
	require.NoError(t, err)

	// time.Sleep(5 * time.Second)

	eventStore, err := esa.NewEventStoreOnDynamoDB(
		dynamodbClient,
		journalTableName, snapshotTableName, journalAidIndexName, snapshotAidIndexName,
		32,
		EventConverter,
		SnapshotConverter,
		esa.WithEventSerializer(NewEventSerializer()),
		esa.WithSnapshotSerializer(NewSnapshotSerializer()))

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
	eventStore := esa.NewEventStoreOnMemory()

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
