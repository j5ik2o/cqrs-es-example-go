package rmu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
)

// recoveryPassFinishedMessage is a metrics contract. CloudWatch Logs metric
// filters in deploy/rmu/cloudwatch-alarms.json read the JSON fields of this
// line, so keep the message and field names stable.
const recoveryPassFinishedMessage = "recovery pass finished"

// RecoveryRunner is the part of RecoveryPass the worker needs.
type RecoveryRunner interface {
	Run(ctx context.Context) (RecoveryPassResult, error)
}

// RecoveryWorker turns an SQS notification or a scheduled invocation into one
// recovery pass. An SQS message body is not read. The durable task in MySQL is
// the source of truth, and the message only wakes the worker.
type RecoveryWorker struct {
	runner RecoveryRunner
}

func NewRecoveryWorker(runner RecoveryRunner) *RecoveryWorker {
	return &RecoveryWorker{runner: runner}
}

// HandleEvent dispatches the Lambda invocation shape. An SQS batch comes from
// the OnFailure queue. An EventBridge schedule, or the empty payload in the
// operations guide, runs one pass. A raw payload without Records is treated as
// a scheduled run so both paths share one function.
func (w *RecoveryWorker) HandleEvent(ctx context.Context, raw json.RawMessage) (any, error) {
	var envelope struct {
		Records []json.RawMessage `json:"Records"`
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return nil, fmt.Errorf("decode Lambda event: %w", err)
		}
	}
	if len(envelope.Records) == 0 {
		return nil, w.HandleScheduledEvent(ctx)
	}
	var event events.SQSEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, fmt.Errorf("decode SQS event: %w", err)
	}
	return w.HandleSQSEvent(ctx, event)
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
	logRecoveryPass(result, err)
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

// HandleScheduledEvent runs one pass from an EventBridge schedule or a manual
// invocation. Unlike the SQS path there is no message to acknowledge, so the
// error is returned. EventBridge retries a failed target and CloudWatch alarms
// see the invocation failure; a permanent data problem stays visible instead
// of being silently deferred.
func (w *RecoveryWorker) HandleScheduledEvent(ctx context.Context) error {
	result, err := w.runner.Run(ctx)
	logRecoveryPass(result, err)
	if err != nil {
		slog.Error("scheduled recovery pass failed", "error", err)
		return err
	}
	return nil
}

// logRecoveryPass writes the completion line that feeds the CloudWatch metric
// filters. deferred and resync_failed are 1/0 so a metric filter can read them
// as numbers.
func logRecoveryPass(result RecoveryPassResult, err error) {
	deferred := 0
	if result.Recovery.Deferred {
		deferred = 1
	}
	resyncFailed := 0
	if result.ResyncFailed {
		resyncFailed = 1
	}
	slog.Info(recoveryPassFinishedMessage,
		"heads", result.Resync.Heads,
		"advanced", result.Resync.Advanced,
		"skipped", result.Resync.Skipped,
		"failed", result.Resync.Failed,
		"completed", result.Recovery.Completed,
		"pending", result.Recovery.Pending,
		"task_failures", result.Recovery.Failed,
		"deferred", deferred,
		"resync_failed", resyncFailed)
	if err != nil {
		slog.Error("recovery pass has errors", "error", err)
	}
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
