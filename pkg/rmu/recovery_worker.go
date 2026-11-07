package rmu

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
)

// RecoveryRunner is the part of RecoveryPass the SQS worker needs.
type RecoveryRunner interface {
	Run(ctx context.Context) (RecoveryPassResult, error)
}

// RecoveryWorker turns an SQS notification into one recovery pass. The message
// body is not read. The durable task in MySQL is the source of truth, and the
// message only wakes the worker.
type RecoveryWorker struct {
	runner RecoveryRunner
}

func NewRecoveryWorker(runner RecoveryRunner) *RecoveryWorker {
	return &RecoveryWorker{runner: runner}
}

// HandleSQSEvent runs one recovery pass for the whole batch. A retryable
// failure returns every message as a batch item failure so SQS redelivers
// them. A permanent failure is acknowledged: the recovery task is already
// stored, and redelivering the same notification cannot fix the data.
//
// The SQS event source mapping must set FunctionResponseTypes to
// ["ReportBatchItemFailures"]; otherwise Lambda deletes every message even
// when this response lists failures.
func (w *RecoveryWorker) HandleSQSEvent(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	result, err := w.runner.Run(ctx)
	slog.Info("recovery pass finished",
		"heads", result.Resync.Heads,
		"advanced", result.Resync.Advanced,
		"skipped", result.Resync.Skipped,
		"failed", result.Resync.Failed,
		"completed", result.Completed)
	if err == nil {
		return events.SQSEventResponse{}, nil
	}
	if !isRetryableAny(err) {
		slog.Error("recovery pass has permanent failures; acknowledged the notification", "error", err)
		return events.SQSEventResponse{}, nil
	}
	failures := make([]events.SQSBatchItemFailure, 0, len(event.Records))
	for _, record := range event.Records {
		failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
	}
	slog.Error("recovery pass failed; returning the batch for redelivery",
		"error", err, "messages", len(failures))
	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}

// isRetryableAny reports whether any error in a joined error is retryable.
// IsRetryable stops at the first ResyncError it finds, which would hide a
// transient recovery failure next to a permanent resync failure.
func isRetryableAny(err error) bool {
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		children := joined.Unwrap()
		if len(children) > 0 {
			for _, child := range children {
				if isRetryableAny(child) {
					return true
				}
			}
			return false
		}
	}
	return IsRetryable(err)
}
