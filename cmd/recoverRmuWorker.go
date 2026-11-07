package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spf13/cobra"
)

// recoverRmuWorkerCmd handles the SQS notifications emitted by the head stream
// mapping's OnFailure destination and the EventBridge schedule that acts as the
// safety net when a notification is lost. The recovery tasks in MySQL are the
// source of truth, so the message body is ignored and one pass covers the
// batch. The logger writes JSON because the CloudWatch metric filters read
// fields from the completion line.
var recoverRmuWorkerCmd = &cobra.Command{
	Use:   "recoverRmuWorker",
	Short: "Replay read model recovery tasks from SQS or a schedule",
	Long:  "Handle an SQS notification or a scheduled event and replay pending read model recovery tasks",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		pass, closeDB, err := newRecoveryPass(context.Background())
		if err != nil {
			panic(err)
		}
		defer closeDB()

		lambda.Start(rmu.NewRecoveryWorker(pass).HandleEvent)
	},
}

func init() {
	rootCmd.AddCommand(recoverRmuWorkerCmd)
}
