package cmd

import (
	"context"
	"cqrs-es-example-go/api"
	repository2 "cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/olivere/env"
	"github.com/spf13/cobra"
)

// writeApiCmd represents the writeApi command
var writeApiCmd = &cobra.Command{
	Use:   "writeApi",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		fmt.Printf("awsRegion = %v\n", awsRegion)
		fmt.Printf("apiHost = %v\n", apiHost)
		fmt.Printf("apiPort = %v\n", apiPort)
		fmt.Printf("journalTableName = %v\n", journalTableName)
		fmt.Printf("snapshotTableName = %v\n", snapshotTableName)
		fmt.Printf("journalAidIndexName = %v\n", journalAidIndexName)
		fmt.Printf("snapshotAidIndexName = %v\n", snapshotAidIndexName)
		fmt.Printf("shardCount = %v\n", shardCount)
		fmt.Printf("awsDynamoDBEndpointUrl = %v\n", awsDynamoDBEndpointUrl)
		fmt.Printf("awsDynamoDBAccessKeyId = %v\n", awsDynamoDBAccessKeyId)
		fmt.Printf("awsDynamoDBSecretKey = %v\n", awsDynamoDBSecretKey)

		var awsCfg aws.Config
		var err error
		if awsDynamoDBEndpointUrl == "" && awsDynamoDBAccessKeyId == "" && awsDynamoDBSecretKey == "" && awsRegion == "" {
			awsCfg, err = config.LoadDefaultConfig(context.Background())
		} else {
			customResolver := aws.EndpointResolverWithOptionsFunc(
				func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           awsDynamoDBEndpointUrl,
						SigningRegion: region,
					}, nil
				})
			awsCfg, err = config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(awsRegion),
				config.WithEndpointResolverWithOptions(customResolver),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsDynamoDBAccessKeyId, awsDynamoDBSecretKey, "x")),
			)
		}
		if err != nil {
			panic(err)
		}

		dynamodbClient := dynamodb.NewFromConfig(awsCfg)

		eventStore, err := esa.NewEventStore(
			dynamodbClient,
			journalTableName,
			snapshotTableName,
			journalAidIndexName,
			snapshotAidIndexName,
			uint64(shardCount),
			repository2.EventConverter,
			repository2.SnapshotConverter,
			esa.WithEventSerializer(repository2.NewEventSerializer()),
			esa.WithSnapshotSerializer(repository2.NewSnapshotSerializer()))
		if err != nil {
			panic(err)
		}

		groupChatRepository := repository2.NewGroupChatRepository(eventStore)
		groupChatController := api.NewGroupChatController(groupChatRepository)

		engine := gin.Default()

		engine.GET("/", api.Index)
		groupChat := engine.Group("/group-chats")
		{
			groupChat.POST("/create", groupChatController.CreateGroupChat)
			groupChat.POST("/rename", groupChatController.RenameGroupChat)
		}
		address := fmt.Sprintf("%s:%d", apiHost, apiPort)
		fmt.Printf("server started at http://%s\n", address)
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
