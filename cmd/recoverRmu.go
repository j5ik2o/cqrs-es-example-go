package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"time"
)

// recoverRmuCmd replays the pending recovery tasks recorded by the RMU. It is
// the worker behind the SQS notification: the message only wakes this command,
// and the event body is read again from the journal.
var recoverRmuCmd = &cobra.Command{
	Use:   "recoverRmu",
	Short: "Replay pending read model recovery tasks",
	Long:  "Replay pending read model recovery tasks from the journal",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		dbUrl := env.String("", "DATABASE_URL")
		if dbUrl == "" {
			panic("DATABASE_URL is required")
		}
		journalTableName := env.String("journal", "PERSISTENCE_JOURNAL_TABLE_NAME")
		headTableName := env.String("head", "PERSISTENCE_HEAD_TABLE_NAME")
		maxEventsPerTransaction := env.Int(1, "PERSISTENCE_MAX_EVENTS_PER_TRANSACTION")

		awsCfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			panic(err)
		}
		dynamodbClient := dynamodb.NewFromConfig(awsCfg)

		db, err := sqlx.Connect("mysql", fmt.Sprintf("%s?parseTime=true", dbUrl))
		if err != nil {
			panic(err.Error())
		}
		defer func(db *sqlx.DB) {
			if db != nil {
				if err := db.Close(); err != nil {
					panic(err.Error())
				}
			}
		}(db)

		dao := rmu.NewGroupChatDaoImpl(db)
		journalReader := rmu.NewDynamoDBJournalReader(dynamodbClient, journalTableName)
		headScanner := rmu.NewDynamoDBHeadScanner(dynamodbClient, headTableName)
		applier := rmu.NewApplier(dao, journalReader, maxEventsPerTransaction)
		recoveryService := rmu.NewRecoveryService(
			rmu.NewMySQLRecoveryTaskStore(db), headScanner, applier)
		resyncer := rmu.NewResyncer(applier, headScanner, recoveryService)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		// A trimmed stream or an OnFailure notification may have hidden heads
		// that never reached this projection. Reconcile every head first, then
		// retry the tasks that a permanent failure recorded.
		result, resyncErr := resyncer.Resync(ctx)
		slog.Info("resync finished", "heads", result.Heads, "advanced", result.Advanced,
			"skipped", result.Skipped, "failed", result.Failed)
		completed, recoveryErr := recoveryService.RecoverPending(ctx)
		slog.Info("recovery run finished", "completed", completed)
		if err := errors.Join(resyncErr, recoveryErr); err != nil {
			slog.Error("some heads or recovery tasks remain unresolved", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(recoverRmuCmd)
}
