package rmu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRecoveryTaskLimit(t *testing.T) {
	require.Equal(t, DefaultRecoveryTaskLimit, NormalizeRecoveryTaskLimit(0))
	require.Equal(t, DefaultRecoveryTaskLimit, NormalizeRecoveryTaskLimit(-5))
	require.Equal(t, 7, NormalizeRecoveryTaskLimit(7))
	require.Equal(t, MaxRecoveryTaskLimit, NormalizeRecoveryTaskLimit(MaxRecoveryTaskLimit+1))
}

func TestRecoveryService_RecoverPendingDefersTasksBeyondLimit(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	first := newReadModelFixture(t)
	second := newReadModelFixture(t)
	third := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		first.aid():  journalEvents(first.aid(), first.create),
		second.aid(): journalEvents(second.aid(), second.create),
		third.aid():  journalEvents(third.aid(), third.create),
	}}
	store := &fakeRecoveryTaskStore{tasks: []RecoveryTask{
		pendingRecoveryTask(first.aid()),
		pendingRecoveryTask(second.aid()),
		pendingRecoveryTask(third.aid()),
	}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier,
		WithRecoveryTaskLimit(2))

	result, err := recovery.RecoverPending(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 2, result.Completed)
	require.Zero(t, result.Failed)
	require.True(t, result.Deferred)
	// The pass reads one task past the limit to learn that work remains.
	require.Equal(t, 1, result.Pending)
	require.Equal(t, int64(1), readModelPosition(t, db, first.aid()))
	require.Equal(t, int64(1), readModelPosition(t, db, second.aid()))
	require.Equal(t, 0, countRows(t, db,
		`SELECT COUNT(*) FROM read_model_positions WHERE aggregate_id = ?`, third.aid()))
}

func TestRecoveryService_RecoverPendingUnderLimitLeavesNothingDeferred(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	first := newReadModelFixture(t)
	second := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		first.aid():  journalEvents(first.aid(), first.create),
		second.aid(): journalEvents(second.aid(), second.create),
	}}
	store := &fakeRecoveryTaskStore{tasks: []RecoveryTask{
		pendingRecoveryTask(first.aid()),
		pendingRecoveryTask(second.aid()),
	}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier,
		WithRecoveryTaskLimit(10))

	result, err := recovery.RecoverPending(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, result.Completed)
	require.False(t, result.Deferred)
	require.Zero(t, result.Pending)
}

func TestRecoveryService_RecoverPendingCountsFailures(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	healthy := newReadModelFixture(t)
	broken := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		healthy.aid(): journalEvents(healthy.aid(), healthy.create),
		broken.aid():  journalEvents(broken.aid(), corruptHeadEvent(1)),
	}}
	store := &fakeRecoveryTaskStore{tasks: []RecoveryTask{
		pendingRecoveryTask(healthy.aid()),
		pendingRecoveryTask(broken.aid()),
	}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)

	result, err := recovery.RecoverPending(ctx)
	require.Error(t, err)
	require.Equal(t, 1, result.Completed)
	require.Equal(t, 1, result.Failed)
	// The failed task stays pending so the next run can retry it.
	require.Equal(t, 1, result.Pending)
}

func TestRecoveryService_FailedTasksDoNotStarveNewerTasks(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	first := newReadModelFixture(t)
	second := newReadModelFixture(t)
	third := newReadModelFixture(t)
	healthy := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		first.aid():   journalEvents(first.aid(), corruptHeadEvent(1)),
		second.aid():  journalEvents(second.aid(), corruptHeadEvent(1)),
		third.aid():   journalEvents(third.aid(), corruptHeadEvent(1)),
		healthy.aid(): journalEvents(healthy.aid(), healthy.create),
	}}
	store := &fakeRecoveryTaskStore{tasks: []RecoveryTask{
		pendingRecoveryTask(first.aid()),
		pendingRecoveryTask(second.aid()),
		pendingRecoveryTask(third.aid()),
		pendingRecoveryTask(healthy.aid()),
	}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier,
		WithRecoveryTaskLimit(2))

	// The first pass fills its limit with permanently failing tasks.
	result, err := recovery.RecoverPending(ctx)
	require.Error(t, err)
	require.Zero(t, result.Completed)
	require.Equal(t, 2, result.Failed)
	require.True(t, result.Deferred)

	// Those failures moved to the back, so the next pass reaches the healthy
	// aggregate instead of retrying the same two tasks.
	result, err = recovery.RecoverPending(ctx)
	require.Error(t, err)
	require.Equal(t, 1, result.Completed)
	require.Equal(t, 1, result.Failed)
	require.Equal(t, int64(1), readModelPosition(t, db, healthy.aid()))
}

func pendingRecoveryTask(aggregateID string) RecoveryTask {
	return RecoveryTask{
		StoreID:           "store-1",
		ProjectionName:    ProjectionName,
		ProjectionVersion: ProjectionVersion,
		AggregateID:       aggregateID,
		SeqNr:             1,
		FailureType:       FailureKindEventPayload,
		ErrorMessage:      "invalid payload",
	}
}
