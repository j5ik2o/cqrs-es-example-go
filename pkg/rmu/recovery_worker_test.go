package rmu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
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

func TestRecoveryWorker_ScheduledPassReturnsRetryableError(t *testing.T) {
	runner := &fakeRecoveryRunner{err: errors.New("connection reset by peer")}
	worker := NewRecoveryWorker(runner)

	err := worker.HandleScheduledEvent(context.Background())

	require.Error(t, err)
	require.Equal(t, 1, runner.calls)
}

func TestRecoveryWorker_ScheduledPassSucceeds(t *testing.T) {
	runner := &fakeRecoveryRunner{}
	worker := NewRecoveryWorker(runner)

	require.NoError(t, worker.HandleScheduledEvent(context.Background()))
	require.Equal(t, 1, runner.calls)
}

func TestRecoveryWorker_HandleEventDispatchesSQSAndSchedule(t *testing.T) {
	ctx := context.Background()

	sqsRunner := &fakeRecoveryRunner{err: errors.New("connection reset by peer")}
	response, err := NewRecoveryWorker(sqsRunner).HandleEvent(ctx,
		json.RawMessage(`{"Records":[{"messageId":"m1"}]}`))
	require.NoError(t, err)
	sqsResponse, ok := response.(events.SQSEventResponse)
	require.True(t, ok)
	require.Len(t, sqsResponse.BatchItemFailures, 1)
	require.Equal(t, "m1", sqsResponse.BatchItemFailures[0].ItemIdentifier)

	// The EventBridge schedule has no Records, so it takes the one-pass path
	// and returns the error for EventBridge to retry.
	scheduledRunner := &fakeRecoveryRunner{err: errors.New("connection reset by peer")}
	_, err = NewRecoveryWorker(scheduledRunner).HandleEvent(ctx,
		json.RawMessage(`{"source":"aws.events","detail-type":"Scheduled Event"}`))
	require.Error(t, err)
	require.Equal(t, 1, scheduledRunner.calls)

	// The operations guide invokes the function with an empty payload.
	emptyRunner := &fakeRecoveryRunner{}
	_, err = NewRecoveryWorker(emptyRunner).HandleEvent(ctx, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.Equal(t, 1, emptyRunner.calls)
}

func TestRecoveryWorker_HandleEventRejectsMalformedPayload(t *testing.T) {
	worker := NewRecoveryWorker(&fakeRecoveryRunner{})

	_, err := worker.HandleEvent(context.Background(), json.RawMessage(`not-json`))

	require.ErrorContains(t, err, "decode Lambda event")
}

func TestRecoveryWorker_LogsMetricFieldsForAlarms(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	runner := &fakeRecoveryRunner{result: RecoveryPassResult{
		Resync: ResyncResult{Heads: 3, Advanced: 1, Failed: 0},
		Recovery: RecoveryPendingResult{
			Completed: 2, Failed: 1, Pending: 4, Deferred: true, Limit: 2,
		},
		ResyncFailed: true,
	}}
	worker := NewRecoveryWorker(runner)

	_, err := worker.HandleSQSEvent(context.Background(), sqsEvent("m1"))
	require.NoError(t, err)

	decoder := json.NewDecoder(&output)
	var line map[string]any
	require.NoError(t, decoder.Decode(&line))
	require.Equal(t, recoveryPassFinishedMessage, line["msg"])
	require.EqualValues(t, 4, line["pending"])
	require.EqualValues(t, 1, line["deferred"])
	require.EqualValues(t, 1, line["task_failures"])
	require.EqualValues(t, 1, line["resync_failed"])
}

func sqsEvent(messageIDs ...string) events.SQSEvent {
	records := make([]events.SQSMessage, 0, len(messageIDs))
	for _, id := range messageIDs {
		records = append(records, events.SQSMessage{MessageId: id})
	}
	return events.SQSEvent{Records: records}
}
