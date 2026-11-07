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
