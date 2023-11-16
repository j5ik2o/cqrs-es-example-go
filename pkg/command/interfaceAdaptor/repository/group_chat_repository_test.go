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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"testing"
)

var (
	journalTableName     = "journal"
	journalAidIndexName  = "journal-aid-index"
	snapshotTableName    = "snapshot"
	snapshotAidIndexName = "snapshot-aid-index"
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

	eventStore, err := esa.NewEventStore(
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
	groupChat, event := domain.NewGroupChat(name, adminId, adminId)
	json, err := json.Marshal(event.ToJSON())
	require.NoError(t, err)
	fmt.Printf("event = %s\n", string(json))
	err = repository.StoreEventWithSnapshot(event, groupChat)
	require.NoError(t, err)

	groupChat2 := repository.FindById(groupChat.GetGroupChatId()).MustGet()
	require.NotNil(t, groupChat2)
	assert.Equal(t, groupChat.GetId(), groupChat2.GetId())
}
