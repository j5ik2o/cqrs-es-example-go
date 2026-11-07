package rmu

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Sleeper waits for a delay or until the context ends. Tests use a fake so the
// retry tests do not spend real time.
type Sleeper interface {
	Sleep(ctx context.Context, delay time.Duration) error
}

type systemSleeper struct{}

func (systemSleeper) Sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// RetryPolicy bounds how long one record may be retried before it is parked.
// MaxElapsed is the wall-clock budget, so a long queue of records cannot push
// a single invocation past the caller's deadline.
type RetryPolicy struct {
	MaxAttempts  int
	MaxElapsed   time.Duration
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Sleeper      Sleeper
	Jitter       func(time.Duration) time.Duration
}

// DefaultRetryPolicy retries a storage failure five times within thirty
// seconds. The delays start at 200ms and cap at 5s, with full jitter.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:  5,
		MaxElapsed:   30 * time.Second,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Sleeper:      systemSleeper{},
		Jitter:       fullJitter,
	}
}

// NoRetryPolicy is used by callers that rely on the outer delivery mechanism
// (Lambda) to retry instead of sleeping inside the invocation.
func NoRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxAttempts: 1}
}

func fullJitter(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(delay) + 1))
}

func (p RetryPolicy) jitter(delay time.Duration) time.Duration {
	if p.Jitter == nil {
		return delay
	}
	return p.Jitter(delay)
}

func (p RetryPolicy) sleeper() Sleeper {
	if p.Sleeper == nil {
		return systemSleeper{}
	}
	return p.Sleeper
}

// Delay returns the jittered delay before the next attempt. attempt starts at 1.
func (p RetryPolicy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if p.InitialDelay <= 0 {
		return 0
	}
	delay := p.InitialDelay
	for i := 1; i < attempt && delay < p.MaxDelay; i++ {
		delay *= 2
	}
	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	return p.jitter(delay)
}

// Do runs op until it succeeds, fails permanently, or the policy is exhausted.
// The returned error wraps the last failure and says how many attempts ran.
func (p RetryPolicy) Do(ctx context.Context, op func() error) error {
	attempts := p.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	startedAt := time.Now()
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		lastErr = op()
		if lastErr == nil {
			return nil
		}
		if !IsRetryable(lastErr) {
			return lastErr
		}
		if attempt == attempts {
			return fmt.Errorf("retry exhausted after %d attempts: %w", attempt, lastErr)
		}
		delay := p.Delay(attempt)
		if p.MaxElapsed > 0 && time.Since(startedAt)+delay > p.MaxElapsed {
			return fmt.Errorf("retry elapsed %s after %d attempts: %w", time.Since(startedAt), attempt, lastErr)
		}
		if err := p.sleeper().Sleep(ctx, delay); err != nil {
			return fmt.Errorf("retry wait: %w", err)
		}
	}
	return lastErr
}

// Backoff is the reader's idle and failure delay. Next grows exponentially up
// to Max and applies full jitter, so a restarted reader does not hammer the
// stream or the database in lock step with another replica.
type Backoff struct {
	Initial time.Duration
	Max     time.Duration
	Jitter  func(time.Duration) time.Duration
	attempt int
}

func NewBackoff(initial, max time.Duration) *Backoff {
	if initial <= 0 {
		initial = 500 * time.Millisecond
	}
	if max <= 0 || max < initial {
		max = 30 * time.Second
	}
	return &Backoff{Initial: initial, Max: max, Jitter: fullJitter}
}

func (b *Backoff) Next() time.Duration {
	delay := b.Initial
	for i := 0; i < b.attempt && delay < b.Max; i++ {
		delay *= 2
	}
	if delay > b.Max {
		delay = b.Max
	}
	b.attempt++
	jitter := b.Jitter
	if jitter == nil {
		jitter = fullJitter
	}
	return jitter(delay)
}

func (b *Backoff) Reset() {
	b.attempt = 0
}
