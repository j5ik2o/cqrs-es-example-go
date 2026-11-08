package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/spf13/cobra"
)

// defaultLocalRecoveryWorkerInterval is how often the local worker replays
// pending recovery tasks. It mirrors the five-minute EventBridge schedule on a
// shorter scale so a local run observes more than one tick.
const defaultLocalRecoveryWorkerInterval = 5 * time.Second

// minLocalRecoveryWorkerInterval stops a bad setting from turning the schedule
// into a busy loop.
const minLocalRecoveryWorkerInterval = 100 * time.Millisecond

// localRecoveryWorkerCmd runs the recovery worker outside Lambda. The startup
// pass models the SQS OnFailure notification; every later tick models the
// EventBridge schedule. Both paths call the same RecoveryWorker handler that
// Lambda invokes, so the JSON completion line is identical to production.
var localRecoveryWorkerCmd = &cobra.Command{
	Use:   "localRecoveryWorker",
	Short: "Run the RMU recovery worker outside Lambda",
	Long:  "Run one full recovery pass at startup and replay pending tasks on a fixed interval",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := runLocalRecoveryWorker(ctx); err != nil {
			slog.Error("local recovery worker stopped", "error", err)
			os.Exit(1)
		}
	},
}

// runLocalRecoveryWorker runs one full pass at startup and then a pending-task
// pass on every tick until the context is cancelled.
func runLocalRecoveryWorker(ctx context.Context) error {
	pass, closeDB, err := newRecoveryPass(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	interval := normalizeLocalRecoveryWorkerInterval(
		durationEnv("RECOVERY_WORKER_INTERVAL", defaultLocalRecoveryWorkerInterval))
	worker := rmu.NewRecoveryWorker(pass)

	slog.Info("local recovery worker started", "interval", interval)

	// The first pass is the SQS OnFailure equivalent: reconcile every head and
	// replay the tasks a permanent failure recorded. A retryable failure
	// returns batch item failures, which SQS would redeliver. Locally the next
	// tick stands in for that redelivery; a scheduled pass would only replay
	// pending tasks and could leave an unscanned head behind forever.
	fullPassCompleted := runFullRecoveryPass(ctx, worker, "local-startup")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if !fullPassCompleted {
				fullPassCompleted = runFullRecoveryPass(ctx, worker, "local-startup-retry")
				if fullPassCompleted {
					slog.Info("startup recovery pass completed after retry")
				}
				continue
			}
			// The schedule replays pending tasks only. It does not scan the
			// head table, so a frequent tick stays cheap.
			if err := worker.HandleScheduledEvent(ctx); err != nil {
				slog.Error("scheduled recovery pass failed", "error", err)
			}
		}
	}
}

// runFullRecoveryPass runs the full recovery path and reports whether it
// completed. Batch item failures mean the notification must be retried, the
// same signal Lambda uses to redeliver an SQS batch.
func runFullRecoveryPass(ctx context.Context, worker *rmu.RecoveryWorker, messageID string) bool {
	response, err := worker.HandleSQSEvent(ctx, dynamodbevents.SQSEvent{
		Records: []dynamodbevents.SQSMessage{{MessageId: messageID}},
	})
	if err != nil {
		slog.Error("full recovery pass failed", "error", err)
		return false
	}
	if len(response.BatchItemFailures) > 0 {
		slog.Warn("full recovery pass reported item failures; retrying on the next tick",
			"failures", len(response.BatchItemFailures))
		return false
	}
	return true
}

// normalizeLocalRecoveryWorkerInterval keeps the local tick inside a range
// that is observable in a test but cannot spin the CPU.
func normalizeLocalRecoveryWorkerInterval(interval time.Duration) time.Duration {
	if interval <= 0 {
		return defaultLocalRecoveryWorkerInterval
	}
	if interval < minLocalRecoveryWorkerInterval {
		return minLocalRecoveryWorkerInterval
	}
	return interval
}

func init() {
	rootCmd.AddCommand(localRecoveryWorkerCmd)
}
