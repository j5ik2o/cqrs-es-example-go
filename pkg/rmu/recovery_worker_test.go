package rmu

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

type fakeRecoveryRunner struct {
	result RecoveryPassResult
	err    error
	calls  int
}

func (f *fakeRecoveryRunner) Run(context.Context) (RecoveryPassResult, error) {
	f.calls++
	return f.result, f.err
}

func TestRecoveryWorker_AcknowledgesSuccessfulPass(t *testing.T) {
	runner := &fakeRecoveryRunner{}
	worker := NewRecoveryWorker(runner)

	response, err := worker.HandleSQSEvent(context.Background(), sqsEvent("m1", "m2"))

	require.NoError(t, err)
	require.Empty(t, response.BatchItemFailures)
	require.Equal(t, 1, runner.calls)
}

func TestRecoveryWorker_RedeliversBatchOnRetryableFailure(t *testing.T) {
	runner := &fakeRecoveryRunner{err: errors.New("connection reset by peer")}
	worker := NewRecoveryWorker(runner)

	response, err := worker.HandleSQSEvent(context.Background(), sqsEvent("m1", "m2"))

	require.NoError(t, err)
	require.Len(t, response.BatchItemFailures, 2)
	require.Equal(t, "m1", response.BatchItemFailures[0].ItemIdentifier)
	require.Equal(t, "m2", response.BatchItemFailures[1].ItemIdentifier)
}

func TestRecoveryWorker_AcknowledgesPermanentFailure(t *testing.T) {
	runner := &fakeRecoveryRunner{err: permanent(FailureKindEventPayload, errors.New("bad payload"))}
	worker := NewRecoveryWorker(runner)

	response, err := worker.HandleSQSEvent(context.Background(), sqsEvent("m1"))

	require.NoError(t, err)
	require.Empty(t, response.BatchItemFailures)
}

func TestRecoveryWorker_AcknowledgesPermanentResyncFailure(t *testing.T) {
	runner := &fakeRecoveryRunner{err: &ResyncError{Permanent: []error{
		permanent(FailureKindEventPayload, errors.New("bad payload")),
	}}}
	worker := NewRecoveryWorker(runner)

	response, err := worker.HandleSQSEvent(context.Background(), sqsEvent("m1"))

	require.NoError(t, err)
	require.Empty(t, response.BatchItemFailures)
}

func TestRecoveryWorker_RetriesWhenAnyJoinedErrorIsRetryable(t *testing.T) {
	// IsRetryable alone stops at the ResyncError and reports the permanent
	// resync failure, hiding the transient database error next to it.
	joined := errors.Join(
		&ResyncError{Permanent: []error{permanent(FailureKindEventPayload, errors.New("bad payload"))}},
		errors.New("connection reset by peer"),
	)
	runner := &fakeRecoveryRunner{err: joined}
	worker := NewRecoveryWorker(runner)

	response, err := worker.HandleSQSEvent(context.Background(), sqsEvent("m1"))

	require.NoError(t, err)
	require.Len(t, response.BatchItemFailures, 1)
	require.Equal(t, "m1", response.BatchItemFailures[0].ItemIdentifier)
}

func TestRecoveryWorker_EmptyBatchDoesNotFail(t *testing.T) {
	runner := &fakeRecoveryRunner{err: errors.New("connection reset by peer")}
	worker := NewRecoveryWorker(runner)

	response, err := worker.HandleSQSEvent(context.Background(), events.SQSEvent{})

	require.NoError(t, err)
	require.Empty(t, response.BatchItemFailures)
	require.Equal(t, 1, runner.calls)
}

func sqsEvent(messageIDs ...string) events.SQSEvent {
	records := make([]events.SQSMessage, 0, len(messageIDs))
	for _, id := range messageIDs {
		records = append(records, events.SQSMessage{MessageId: id})
	}
	return events.SQSEvent{Records: records}
}
