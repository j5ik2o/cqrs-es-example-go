package cmd

import (
	"context"
	commandgraphql "cqrs-es-example-go/pkg/command/interfaceAdaptor/graphql"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"cqrs-es-example-go/pkg/command/processor"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
	"github.com/olivere/env"
	"github.com/spf13/cobra"
	"log/slog"
	"net/http"
	"os"
)

const writeApiDefaultPort = 28080

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
		apiPort := env.Int(writeApiDefaultPort, "API_PORT")
		journalTableName := env.String("journal", "PERSISTENCE_JOURNAL_TABLE_NAME")
		snapshotTableName := env.String("snapshot", "PERSISTENCE_SNAPSHOT_TABLE_NAME")
		journalAidIndexName := env.String("journal-aid-index", "PERSISTENCE_JOURNAL_AID_INDEX_NAME")
		snapshotAidIndexName := env.String("snapshot-aid-index", "PERSISTENCE_SNAPSHOT_AID_INDEX_NAME")
		shardCount := env.Int64(10, "PERSISTENCE_SHARD_COUNT")
		awsDynamoDBEndpointUrl := env.String("", "AWS_DYNAMODB_ENDPOINT_URL")
		awsDynamoDBAccessKeyId := env.String("", "AWS_DYNAMODB_ACCESS_KEY_ID")
		awsDynamoDBSecretKey := env.String("", "AWS_DYNAMODB_SECRET_ACCESS_KEY")

		slog.Info(fmt.Sprintf("AWS_REGION = %v", awsRegion))
		slog.Info(fmt.Sprintf("API_HOST = %v", apiHost))
		slog.Info(fmt.Sprintf("API_PORT = %v", apiPort))
		slog.Info(fmt.Sprintf("PERSISTENCE_JOURNAL_TABLE_NAME = %v", journalTableName))
		slog.Info(fmt.Sprintf("PERSISTENCE_SNAPSHOT_TABLE_NAME = %v", snapshotTableName))
		slog.Info(fmt.Sprintf("PERSISTENCE_JOURNAL_AID_INDEX_NAME = %v", journalAidIndexName))
		slog.Info(fmt.Sprintf("PERSISTENCE_SNAPSHOT_AID_INDEX_NAME = %v", snapshotAidIndexName))
		slog.Info(fmt.Sprintf("PERSISTENCE_SHARD_COUNT = %v", shardCount))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_ENDPOINT_URL = %v", awsDynamoDBEndpointUrl))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_ACCESS_KEY_ID = %v", awsDynamoDBAccessKeyId))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_SECRET_ACCESS_KEY = %v", awsDynamoDBSecretKey))

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

		groupChatRepository, err := repository.NewGroupChatRepository(eventStore, repository.WithRetention(100))
		if err != nil {
			panic(err)
		}
		groupChatCommandProcessor := processor.NewGroupChatCommandProcessor(&groupChatRepository)

		srv := handler.NewDefaultServer(commandgraphql.NewExecutableSchema(commandgraphql.Config{Resolvers: commandgraphql.NewResolver(groupChatCommandProcessor)}))

		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		http.Handle("/query", srv)

		endpoint := fmt.Sprintf("%s:%d", apiHost, apiPort)
		slog.Info(fmt.Sprintf("connect to http://%s/ for GraphQL playground", endpoint))
		err = http.ListenAndServe(endpoint, nil)
		if err != nil {
			slog.Error("failed to start server", "error", err.Error())
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
