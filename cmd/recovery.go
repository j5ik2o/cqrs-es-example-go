package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
)

// newRecoveryPass builds the resync and recovery pipeline shared by the
// one-shot recoverRmu command and the SQS-triggered recoverRmuWorker. The
// returned cleanup closes the database when the caller is a short-lived
// process.
func newRecoveryPass(ctx context.Context) (*rmu.RecoveryPass, func(), error) {
	dbUrl := env.String("", "DATABASE_URL")
	if dbUrl == "" {
		return nil, nil, fmt.Errorf("DATABASE_URL is required")
	}
	journalTableName := env.String("journal", "PERSISTENCE_JOURNAL_TABLE_NAME")
	headTableName := env.String("head", "PERSISTENCE_HEAD_TABLE_NAME")
	maxEventsPerTransaction := env.Int(1, "PERSISTENCE_MAX_EVENTS_PER_TRANSACTION")

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("load AWS config: %w", err)
	}
	dynamodbClient := dynamodb.NewFromConfig(awsCfg)

	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s?parseTime=true", dbUrl))
	if err != nil {
		return nil, nil, fmt.Errorf("connect to MySQL: %w", err)
	}
	closeDB := func() {
		if err := db.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}

	dao := rmu.NewGroupChatDaoImpl(db)
	journalReader := rmu.NewDynamoDBJournalReader(dynamodbClient, journalTableName)
	headScanner := rmu.NewDynamoDBHeadScanner(dynamodbClient, headTableName)
	applier := rmu.NewApplier(dao, journalReader, maxEventsPerTransaction)
	recoveryService := rmu.NewRecoveryService(
		rmu.NewMySQLRecoveryTaskStore(db), headScanner, applier)
	resyncer := rmu.NewResyncer(applier, headScanner, recoveryService)
	return rmu.NewRecoveryPass(resyncer, recoveryService), closeDB, nil
}
