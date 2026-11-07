package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// recoverRmuCmd replays the pending recovery tasks recorded by the RMU. It is
// the manual counterpart of the SQS worker: the message only wakes the worker,
// and the event body is read again from the journal.
var recoverRmuCmd = &cobra.Command{
	Use:   "recoverRmu",
	Short: "Replay pending read model recovery tasks",
	Long:  "Replay pending read model recovery tasks from the journal",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		pass, closeDB, err := newRecoveryPass(context.Background())
		if err != nil {
			panic(err)
		}
		defer closeDB()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		// A trimmed stream or an OnFailure notification may have hidden heads
		// that never reached this projection. Reconcile every head first, then
		// retry the tasks that a permanent failure recorded.
		result, runErr := pass.Run(ctx)
		slog.Info("recovery run finished",
			"heads", result.Resync.Heads,
			"advanced", result.Resync.Advanced,
			"skipped", result.Resync.Skipped,
			"failed", result.Resync.Failed,
			"completed", result.Completed)
		if runErr != nil {
			slog.Error("some heads or recovery tasks remain unresolved", "error", runErr)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(recoverRmuCmd)
}
