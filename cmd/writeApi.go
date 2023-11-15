package cmd

import (
	"context"
	"cqrs-es-example-go/api"
	repository2 "cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"github.com/aws/aws-sdk-go-v2/config"
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
		}
		err = engine.Run()
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
