package main

import (
	"context"
	controller "cqrs-es-example-go/interfaceAdapter/ctrl"
	"cqrs-es-example-go/repository"
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

	eventStore, err := esa.NewEventStore(
		dynamodbClient,
		journalTableName,
		snapshotTableName,
		journalAidIndexName,
		snapshotAidIndexName,
		uint64(shardCount),
		repository.EventConverter,
		repository.SnapshotConverter,
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
