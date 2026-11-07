package rmu

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetryPolicy_DoRetriesUntilSuccess(t *testing.T) {
	sleeper := &fakeSleeper{}
	policy := RetryPolicy{
		MaxAttempts:  3,
		MaxElapsed:   time.Minute,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     20 * time.Millisecond,
		Sleeper:      sleeper,
		Jitter:       func(d time.Duration) time.Duration { return d },
	}
	attempts := 0
	err := policy.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary outage")
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, attempts)
	require.Equal(t, []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}, sleeper.delays)
}

func TestRetryPolicy_DoStopsAtMaxAttempts(t *testing.T) {
	sleeper := &fakeSleeper{}
	policy := RetryPolicy{
		MaxAttempts:  3,
		MaxElapsed:   time.Minute,
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
		Sleeper:      sleeper,
		Jitter:       func(d time.Duration) time.Duration { return d },
	}
	attempts := 0
	err := policy.Do(context.Background(), func() error {
		attempts++
		return errors.New("database unavailable")
	})
	require.ErrorContains(t, err, "retry exhausted after 3 attempts")
	require.Equal(t, 3, attempts)
	require.Len(t, sleeper.delays, 2)
}

func TestRetryPolicy_DoStopsAtMaxElapsed(t *testing.T) {
	sleeper := &fakeSleeper{}
	policy := RetryPolicy{
		MaxAttempts:  100,
		MaxElapsed:   time.Nanosecond,
		InitialDelay: time.Hour,
		MaxDelay:     time.Hour,
		Sleeper:      sleeper,
		Jitter:       func(d time.Duration) time.Duration { return d },
	}
	err := policy.Do(context.Background(), func() error {
		return errors.New("database unavailable")
	})
	require.ErrorContains(t, err, "retry elapsed")
	require.Empty(t, sleeper.delays)
}

func TestRetryPolicy_DoesNotRetryPermanentFailure(t *testing.T) {
	sleeper := &fakeSleeper{}
	policy := RetryPolicy{
		MaxAttempts:  5,
		MaxElapsed:   time.Minute,
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
		Sleeper:      sleeper,
	}
	attempts := 0
	err := policy.Do(context.Background(), func() error {
		attempts++
		return permanent(FailureKindEventPayload, errors.New("bad payload"))
	})
	require.Error(t, err)
	require.Equal(t, 1, attempts)
	require.Empty(t, sleeper.delays)
	kind, ok := ClassifyFailure(err)
	require.True(t, ok)
	require.Equal(t, FailureKindEventPayload, kind)
}

func TestBackoff_GrowsCapsAndResets(t *testing.T) {
	backoff := NewBackoff(100*time.Millisecond, time.Second)
	backoff.Jitter = func(d time.Duration) time.Duration { return d }
	require.Equal(t, 100*time.Millisecond, backoff.Next())
	require.Equal(t, 200*time.Millisecond, backoff.Next())
	require.Equal(t, 400*time.Millisecond, backoff.Next())
	require.Equal(t, 800*time.Millisecond, backoff.Next())
	require.Equal(t, time.Second, backoff.Next())
	require.Equal(t, time.Second, backoff.Next())

	backoff.Reset()
	require.Equal(t, 100*time.Millisecond, backoff.Next())
}

func TestBackoff_FullJitterStaysWithinDelay(t *testing.T) {
	backoff := NewBackoff(100*time.Millisecond, time.Second)
	for i := 0; i < 100; i++ {
		delay := backoff.Next()
		require.GreaterOrEqual(t, delay, time.Duration(0))
		require.LessOrEqual(t, delay, time.Second)
		if i > 10 {
			break
		}
	}
}

func TestClassifyFailure_RecoveryTaskWriteFailureIsRetryable(t *testing.T) {
	err := &RecoveryRecordError{
		ApplyErr:  permanent(FailureKindEventPayload, errors.New("bad payload")),
		RecordErr: errors.New("database unavailable"),
	}
	_, ok := ClassifyFailure(err)
	require.False(t, ok)
	require.True(t, IsRetryable(err))
}

func TestClassifyFailure_MySQLConstraintIsPermanent(t *testing.T) {
	kind, ok := ClassifyFailure(errors.New("plain error"))
	require.False(t, ok)
	require.Equal(t, FailureKindUnknown, kind)

	require.True(t, IsRetryable(errors.New("plain error")))
	require.False(t, IsRetryable(context.Canceled))
}
