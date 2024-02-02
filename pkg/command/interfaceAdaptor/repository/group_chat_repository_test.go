package repository

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/models"
	"encoding/json"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/j5ik2o/event-store-adapter-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	container, err := CreateContainer(ctx)
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
	repository := NewGroupChatRepository(eventStore)
	adminId := models.NewUserAccountId()
	name, err := models.NewGroupChatName("test").Get()
	if err != nil {
		t.Fatal(err)
	}
	groupChat, event := domain.NewGroupChat(name, adminId)
	jsonObj, err := json.Marshal(event.ToJSON())
	require.NoError(t, err)
	fmt.Printf("event = %s\n", string(jsonObj))
	err = repository.StoreEventWithSnapshot(event, &groupChat)
	require.NoError(t, err)

	groupChat2 := repository.FindById(groupChat.GetGroupChatId()).MustGet()
	require.NotNil(t, groupChat2)
	assert.Equal(t, groupChat.GetId(), groupChat2.GetId())
}

func Test_GroupChatRepository_OnMemory_FindById(t *testing.T) {
	eventStore := esa.NewEventStoreOnMemory()

	repository := NewGroupChatRepository(eventStore)
	adminId := models.NewUserAccountId()
	name, err := models.NewGroupChatName("test").Get()
	if err != nil {
		t.Fatal(err)
	}
	groupChat, event := domain.NewGroupChat(name, adminId)
	jsonObj, err := json.Marshal(event.ToJSON())
	require.NoError(t, err)
	fmt.Printf("event = %s\n", string(jsonObj))
	err = repository.StoreEventWithSnapshot(event, &groupChat)
	require.NoError(t, err)

	groupChat2 := repository.FindById(groupChat.GetGroupChatId()).MustGet()
	require.NotNil(t, groupChat2)
	assert.Equal(t, groupChat.GetId(), groupChat2.GetId())
}
