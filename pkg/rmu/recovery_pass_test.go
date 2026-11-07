package rmu

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoveryPass_RunsRecoveryEvenWhenResyncFails(t *testing.T) {
	// A transient scan failure must not stop the pending tasks from being
	// replayed; both errors reach the caller.
	resyncer := NewResyncer(nil, &fakeHeadScanner{err: errors.New("connection reset by peer")}, nil)
	recovery := NewRecoveryService(
		&fakeRecoveryTaskStore{listErr: errors.New("recovery table is unavailable")},
		staticStoreID{"store-1"}, nil)
	pass := NewRecoveryPass(resyncer, recovery)

	_, err := pass.Run(context.Background())

	require.ErrorContains(t, err, "resync scan failed")
	require.ErrorContains(t, err, "list recovery tasks")
}

func TestRecoveryPass_ScheduledRunSkipsHeadScan(t *testing.T) {
	// The schedule only replays pending tasks. A head scan on every tick would
	// read the whole table for no reason.
	scanner := &fakeHeadScanner{err: errors.New("head scan must not run")}
	resyncer := NewResyncer(nil, scanner, nil)
	recovery := NewRecoveryService(&fakeRecoveryTaskStore{}, staticStoreID{"store-1"}, nil)
	pass := NewRecoveryPass(resyncer, recovery)

	result, err := pass.RunScheduled(context.Background())

	require.NoError(t, err)
	require.False(t, result.ResyncFailed)
	require.Zero(t, result.Resync.Heads)
	require.Zero(t, result.Recovery.Completed)
}
