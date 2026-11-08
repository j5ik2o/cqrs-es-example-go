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

func TestRecoveryPass_ScheduledRunRetriesPendingTaskOnNextPass(t *testing.T) {
	// A scheduled pass that cannot replay a task leaves it pending. The next
	// tick must retry that same task and finish the replay, so a transient
	// failure during one schedule run does not wait for another notification.
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{
		events: map[string][]JournalEvent{
			fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember),
		},
		errs: []error{errors.New("connection reset by peer")},
	}
	store := &fakeRecoveryTaskStore{tasks: []RecoveryTask{{
		StoreID:           "store-1",
		ProjectionName:    ProjectionName,
		ProjectionVersion: ProjectionVersion,
		AggregateID:       fixture.aid(),
		SeqNr:             2,
		FailureType:       FailureKindJournalGap,
		ErrorMessage:      "simulated transient failure",
	}}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	pass := NewRecoveryPass(NewResyncer(applier, nil, recovery), recovery)

	result, err := pass.RunScheduled(ctx)
	require.ErrorContains(t, err, "read journal")
	require.Zero(t, result.Recovery.Completed)
	require.Equal(t, 1, result.Recovery.Failed)
	require.Equal(t, 1, result.Recovery.Pending)
	require.Equal(t, 0, countRows(t, db,
		`SELECT COUNT(*) FROM read_model_positions WHERE aggregate_id = ?`, fixture.aid()))

	result, err = pass.RunScheduled(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, result.Recovery.Completed)
	require.Zero(t, result.Recovery.Failed)
	require.Zero(t, result.Recovery.Pending)
	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
}
