package main

import (
	"context"
	controller "cqrs-es-example-go/interfaceAdapter/ctrl"
	"cqrs-es-example-go/repository"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/olivere/env"
)

func main() {
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	dynamodbClient := dynamodb.NewFromConfig(awsCfg)
	journalTableName := env.String("journal", "JOURNAL_TABLE_NAME")
	snapshotTableName := env.String("snapshot", "SNAPSHOT_TABLE_NAME")
	journalAidIndexName := env.String("journal-aid-index", "JOURNAL_AID_INDEX_NAME")
	snapshotAidIndexName := env.String("snapshot-aid-index", "SNAPSHOT_AID_INDEX_NAME")
	shardCount := env.Int64(10, "SHARD_COUNT")

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
