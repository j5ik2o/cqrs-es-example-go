package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeLocalRecoveryWorkerInterval(t *testing.T) {
	require.Equal(t, defaultLocalRecoveryWorkerInterval,
		normalizeLocalRecoveryWorkerInterval(0))
	require.Equal(t, defaultLocalRecoveryWorkerInterval,
		normalizeLocalRecoveryWorkerInterval(-time.Second))
	require.Equal(t, minLocalRecoveryWorkerInterval,
		normalizeLocalRecoveryWorkerInterval(time.Millisecond))
	require.Equal(t, 2*time.Second,
		normalizeLocalRecoveryWorkerInterval(2*time.Second))
}

func TestRunFullRecoveryPassRetriesOnRetryableFailure(t *testing.T) {
	worker := rmu.NewRecoveryWorker(&fakeRecoveryRunner{err: errors.New("storage unavailable")})
	require.False(t, runFullRecoveryPass(context.Background(), worker, "local-startup"))
}

func TestRunFullRecoveryPassCompletesWithoutFailures(t *testing.T) {
	worker := rmu.NewRecoveryWorker(&fakeRecoveryRunner{})
	require.True(t, runFullRecoveryPass(context.Background(), worker, "local-startup"))
}

// fakeRecoveryRunner stands in for the RecoveryPass so the local worker's
// retry decision can be tested without MySQL or DynamoDB.
type fakeRecoveryRunner struct {
	err error
}

func (f *fakeRecoveryRunner) Run(context.Context) (rmu.RecoveryPassResult, error) {
	return rmu.RecoveryPassResult{}, f.err
}

func (f *fakeRecoveryRunner) RunScheduled(context.Context) (rmu.RecoveryPassResult, error) {
	return rmu.RecoveryPassResult{}, f.err
}
