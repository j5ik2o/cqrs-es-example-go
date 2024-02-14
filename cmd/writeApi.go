package cmd

import (
	"context"
	"cqrs-es-example-go/api"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"cqrs-es-example-go/pkg/command/useCase"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/olivere/env"
	sloggin "github.com/samber/slog-gin"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

// writeApiCmd represents the writeApi command
var writeApiCmd = &cobra.Command{
	Use:   "writeApi",
	Short: "Write API",
	Long:  "Write API",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		awsRegion := env.String("", "AWS_REGION")
		apiHost := env.String("0.0.0.0", "API_HOST")
		apiPort := env.Int(8080, "API_PORT")
		journalTableName := env.String("journal", "PERSISTENCE_JOURNAL_TABLE_NAME")
		snapshotTableName := env.String("snapshot", "PERSISTENCE_SNAPSHOT_TABLE_NAME")
		journalAidIndexName := env.String("journal-aid-index", "PERSISTENCE_JOURNAL_AID_INDEX_NAME")
		snapshotAidIndexName := env.String("snapshot-aid-index", "PERSISTENCE_SNAPSHOT_AID_INDEX_NAME")
		shardCount := env.Int64(10, "PERSISTENCE_SHARD_COUNT")
		awsDynamoDBEndpointUrl := env.String("", "AWS_DYNAMODB_ENDPOINT_URL")
		awsDynamoDBAccessKeyId := env.String("", "AWS_DYNAMODB_ACCESS_KEY_ID")
		awsDynamoDBSecretKey := env.String("", "AWS_DYNAMODB_SECRET_ACCESS_KEY")

		slog.Info(fmt.Sprintf("awsRegion = %v", awsRegion))
		slog.Info(fmt.Sprintf("apiHost = %v", apiHost))
		slog.Info(fmt.Sprintf("apiPort = %v", apiPort))
		slog.Info(fmt.Sprintf("journalTableName = %v", journalTableName))
		slog.Info(fmt.Sprintf("snapshotTableName = %v", snapshotTableName))
		slog.Info(fmt.Sprintf("journalAidIndexName = %v", journalAidIndexName))
		slog.Info(fmt.Sprintf("snapshotAidIndexName = %v", snapshotAidIndexName))
		slog.Info(fmt.Sprintf("shardCount = %v", shardCount))
		slog.Info(fmt.Sprintf("awsDynamoDBEndpointUrl = %v", awsDynamoDBEndpointUrl))
		slog.Info(fmt.Sprintf("awsDynamoDBAccessKeyId = %v", awsDynamoDBAccessKeyId))
		slog.Info(fmt.Sprintf("awsDynamoDBSecretKey = %v", awsDynamoDBSecretKey))

		var awsCfg aws.Config
		var err error
		if awsDynamoDBEndpointUrl == "" && awsDynamoDBAccessKeyId == "" && awsDynamoDBSecretKey == "" && awsRegion == "" {
			awsCfg, err = config.LoadDefaultConfig(context.Background())
		} else {
			if awsRegion == "" {
				panic("AWS_REGION is required")
			}
			if awsDynamoDBEndpointUrl == "" {
				panic("AWS_DYNAMODB_ENDPOINT_URL is required")
			}
			if awsDynamoDBAccessKeyId == "" {
				panic("AWS_DYNAMODB_ACCESS_KEY_ID is required")
			}
			if awsDynamoDBSecretKey == "" {
				panic("AWS_DYNAMODB_SECRET_ACCESS_KEY is required")
			}
			customResolver := aws.EndpointResolverWithOptionsFunc(
				func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           awsDynamoDBEndpointUrl,
						SigningRegion: region,
					}, nil
				})
			awsCfg, err = config.LoadDefaultConfig(context.Background(),
				config.WithRegion(awsRegion),
				config.WithEndpointResolverWithOptions(customResolver),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsDynamoDBAccessKeyId, awsDynamoDBSecretKey, "x")),
			)
		}
		if err != nil {
			panic(err)
		}

		dynamodbClient := dynamodb.NewFromConfig(awsCfg)

		eventStore, err := esa.NewEventStoreOnDynamoDB(
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
		if err != nil {
			panic(err)
		}

		groupChatRepository := repository.NewGroupChatRepository(eventStore)
		groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
		groupChatController := api.NewGroupChatController(groupChatCommandProcessor)

		engine := gin.New()
		engine.Use(sloggin.New(logger))
		engine.Use(gin.Recovery())

		engine.GET("/", api.Index)
		groupChat := engine.Group("/group-chats")
		{
			groupChat.POST("/create", groupChatController.CreateGroupChat)
			groupChat.POST("/delete", groupChatController.DeleteGroupChat)
			groupChat.POST("/rename", groupChatController.RenameGroupChat)
			groupChat.POST("/add-member", groupChatController.AddMember)
			groupChat.POST("/remove-member", groupChatController.RemoveMember)
			groupChat.POST("/post-message", groupChatController.PostMessage)
			groupChat.POST("/delete-message", groupChatController.DeleteMessage)
		}
		address := fmt.Sprintf("%s:%d", apiHost, apiPort)
		slog.Info(fmt.Sprintf("server started at http://%s", address))
		err = engine.Run(address)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(writeApiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// writeApiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// writeApiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
