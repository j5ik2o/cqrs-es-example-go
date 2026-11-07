package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestUpdateReadModel_PermanentFailureRecordsRecoveryTask(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	corrupt := corruptHeadEvent(2)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, corrupt),
	}}
	store := &fakeRecoveryTaskStore{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal,
		WithApplier(applier), WithRecoveryService(recovery))

	err := updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "INSERT", fixture.aid(), fixture.create),
		headRecord(t, "MODIFY", fixture.aid(), corrupt),
	}})
	require.Error(t, err)

	require.Equal(t, int64(1), readModelPosition(t, db, fixture.aid()))
	require.Len(t, store.tasks, 1)
	task := store.tasks[0]
	require.Equal(t, "store-1", task.StoreID)
	require.Equal(t, ProjectionName, task.ProjectionName)
	require.Equal(t, ProjectionVersion, task.ProjectionVersion)
	require.Equal(t, fixture.aid(), task.AggregateID)
	require.Equal(t, int64(2), task.SeqNr)
	require.Equal(t, FailureKindEventPayload, task.FailureType)
}

func TestUpdateReadModel_TransientFailureRetriesThenSucceeds(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{
		events: map[string][]JournalEvent{
			fixture.aid(): journalEvents(fixture.aid(), fixture.create),
		},
		errs: []error{errors.New("connection reset by peer")},
	}
	sleeper := &fakeSleeper{}
	policy := RetryPolicy{
		MaxAttempts:  3,
		MaxElapsed:   time.Minute,
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
		Sleeper:      sleeper,
		Jitter:       func(d time.Duration) time.Duration { return d },
	}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal, WithRetryPolicy(policy))

	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "MODIFY", fixture.aid(), fixture.addMember),
	}}))

	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
	require.Len(t, sleeper.delays, 1)
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestUpdateReadModel_RetryExhaustionLeavesReadModelUnchanged(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{
		events: map[string][]JournalEvent{fixture.aid(): journalEvents(fixture.aid(), fixture.create)},
		errs:   []error{errors.New("connection refused"), errors.New("connection refused"), errors.New("connection refused")},
	}
	sleeper := &fakeSleeper{}
	policy := RetryPolicy{
		MaxAttempts:  3,
		MaxElapsed:   time.Minute,
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
		Sleeper:      sleeper,
		Jitter:       func(d time.Duration) time.Duration { return d },
	}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal, WithRetryPolicy(policy))

	err := updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "MODIFY", fixture.aid(), fixture.addMember),
	}})
	require.Error(t, err)

	// The backoff ran twice before the third attempt failed, and nothing was
	// written: the position and the read model are both untouched.
	require.Len(t, sleeper.delays, 2)
	require.Equal(t, 0, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 0, countRows(t, db, `SELECT COUNT(*) FROM read_model_positions`))
}

func TestUpdateReadModel_RecoveryStoreWriteFailureIsReturned(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	corrupt := corruptHeadEvent(2)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, corrupt),
	}}
	store := &fakeRecoveryTaskStore{recordErr: errors.New("recovery table is unavailable")}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal,
		WithApplier(applier), WithRecoveryService(recovery))

	err := updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "INSERT", fixture.aid(), fixture.create),
		headRecord(t, "MODIFY", fixture.aid(), corrupt),
	}})
	require.ErrorContains(t, err, "record recovery task")
	require.Equal(t, int64(1), readModelPosition(t, db, fixture.aid()))
}

func TestUpdateReadModel_ConcurrentApply(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create),
	}}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)
	event := dynamodbevents.DynamoDBEvent{Records: []dynamodbevents.DynamoDBEventRecord{
		headRecord(t, "INSERT", fixture.aid(), fixture.create),
	}}

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- updater.UpdateReadModel(ctx, event)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	require.Equal(t, int64(1), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestGroupChatDao_ConcurrentPositionRowCreation(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	dao := NewGroupChatDaoImpl(db)

	firstGroupChatID := models.NewGroupChatId()
	secondGroupChatID := models.NewGroupChatId()
	aggregateIDs := []string{
		firstGroupChatID.AsString(),
		secondGroupChatID.AsString(),
	}
	start := make(chan struct{})
	errs := make(chan error, len(aggregateIDs))
	var wg sync.WaitGroup
	for _, aggregateID := range aggregateIDs {
		aggregateID := aggregateID
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- dao.Process(ctx, aggregateID, func(_ ReadModelWriter, current int64) (int64, error) {
				return current, nil
			})
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	require.Equal(t, 2, countRows(t, db,
		`SELECT COUNT(*) FROM read_model_positions WHERE aggregate_id IN (?, ?)`,
		aggregateIDs[0], aggregateIDs[1]))
}

func TestResync_FillsLagFromJournalAfterLongStop(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember, fixture.rename),
	}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal, WithApplier(applier))

	// The projection stopped after seq_nr 1 while the aggregate advanced to 3.
	require.NoError(t, updater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "INSERT", fixture.aid(), fixture.create),
		},
	}))
	require.Equal(t, int64(1), readModelPosition(t, db, fixture.aid()))

	scanner := &fakeHeadScanner{heads: []Head{
		{AggregateID: fixture.aid(), TypeName: models.GroupChatIdPrefix, SeqNr: 3},
	}}
	resyncer := NewResyncer(applier, scanner, nil)
	result, err := resyncer.Resync(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, result.Heads)
	require.Equal(t, 1, result.Advanced)

	require.Equal(t, int64(3), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, "renamed", readModelName(t, db, fixture.aid()))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
}

func TestResync_SkipsConfigAndForeignHeads(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	scanner := &fakeHeadScanner{heads: []Head{
		{AggregateID: configAggregateID, TypeName: models.GroupChatIdPrefix, SeqNr: 1},
		{AggregateID: "UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z", TypeName: "UserAccount", SeqNr: 1},
		{AggregateID: fixture.aid(), TypeName: models.GroupChatIdPrefix, SeqNr: 0},
	}}
	resyncer := NewResyncer(applier, scanner, nil)
	result, err := resyncer.Resync(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, result.Heads)
	require.Equal(t, 2, result.Skipped)
	require.Equal(t, 0, countRows(t, db, `SELECT COUNT(*) FROM read_model_positions`))
}

func TestResync_PermanentFailureRecordsTaskAndCompletesOtherAggregates(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	broken := newReadModelFixture(t)
	healthy := newReadModelFixture(t)
	corrupt := corruptHeadEvent(2)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		broken.aid():  journalEvents(broken.aid(), broken.create, corrupt),
		healthy.aid(): journalEvents(healthy.aid(), healthy.create, healthy.addMember),
	}}
	store := &fakeRecoveryTaskStore{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	scanner := &fakeHeadScanner{heads: []Head{
		{AggregateID: broken.aid(), TypeName: models.GroupChatIdPrefix, SeqNr: 2},
		{AggregateID: healthy.aid(), TypeName: models.GroupChatIdPrefix, SeqNr: 2},
	}}
	resyncer := NewResyncer(applier, scanner, recovery)

	_, err := resyncer.Resync(ctx)
	require.Error(t, err)
	require.Len(t, store.tasks, 1)
	require.Equal(t, broken.aid(), store.tasks[0].AggregateID)
	require.Equal(t, int64(2), store.tasks[0].SeqNr)

	// seq_nr 1 committed before seq_nr 2 failed, but the position never passes
	// the event that is still unapplied.
	require.Equal(t, int64(1), readModelPosition(t, db, broken.aid()))
	require.Equal(t, int64(2), readModelPosition(t, db, healthy.aid()))
}

func TestRecoveryService_RecoverIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember),
	}}
	store := &fakeRecoveryTaskStore{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	task := RecoveryTask{
		StoreID:           "store-1",
		ProjectionName:    ProjectionName,
		ProjectionVersion: ProjectionVersion,
		AggregateID:       fixture.aid(),
		SeqNr:             2,
		FailureType:       FailureKindEventPayload,
	}

	require.NoError(t, recovery.Recover(ctx, task))
	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))

	// A redelivered task finds the position at the target and only marks the
	// task complete; the read model is not applied twice.
	require.NoError(t, recovery.Recover(ctx, task))
	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
	require.Equal(t, 1, countRows(t, db, `SELECT COUNT(*) FROM group_chats`))
	require.Equal(t, 2, countRows(t, db, `SELECT COUNT(*) FROM members`))
	require.Len(t, store.completed, 2)
}

func TestRecoveryService_RecoverFailsWhenPositionDoesNotReachTarget(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create),
	}}
	store := &fakeRecoveryTaskStore{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	task := RecoveryTask{
		StoreID: "store-1", ProjectionName: ProjectionName, ProjectionVersion: ProjectionVersion,
		AggregateID: fixture.aid(), SeqNr: 3,
	}

	err := recovery.Recover(ctx, task)
	require.Error(t, err)
	require.Empty(t, store.completed)
	require.Equal(t, 0, countRows(t, db,
		`SELECT COUNT(*) FROM read_model_positions WHERE aggregate_id = ?`, fixture.aid()))
}

func TestMySQLRecoveryTaskStore_RecordIsIdempotentAndCompletes(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	store := NewMySQLRecoveryTaskStore(db)
	task := RecoveryTask{
		StoreID:           "store-1",
		ProjectionName:    ProjectionName,
		ProjectionVersion: ProjectionVersion,
		AggregateID:       "GroupChat-01H42K4ABWQ5V2XQEP3A48VE0Z",
		SeqNr:             2,
		FailureType:       FailureKindEventPayload,
		ErrorMessage:      "invalid payload",
	}

	require.NoError(t, store.Record(ctx, task))
	task.ErrorMessage = "still invalid"
	require.NoError(t, store.Record(ctx, task))

	filter := RecoveryTaskFilter{
		StoreID: task.StoreID, ProjectionName: task.ProjectionName, ProjectionVersion: task.ProjectionVersion,
	}
	pending, err := store.ListPending(ctx, filter)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, "still invalid", pending[0].ErrorMessage)
	require.Equal(t, FailureKindEventPayload, pending[0].FailureType)

	require.NoError(t, store.MarkCompleted(ctx, task.StoreID, task.ProjectionName, task.AggregateID, task.SeqNr))
	pending, err = store.ListPending(ctx, filter)
	require.NoError(t, err)
	require.Empty(t, pending)
	count, err := store.CountPending(ctx, filter)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestMySQLRecoveryTaskStore_ListPendingHonoursFilterAndLimit(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	store := NewMySQLRecoveryTaskStore(db)
	filter := RecoveryTaskFilter{
		StoreID: "store-1", ProjectionName: ProjectionName, ProjectionVersion: ProjectionVersion,
	}
	for _, aid := range []string{
		"GroupChat-01H42K4ABWQ5V2XQEP3A48VE01",
		"GroupChat-01H42K4ABWQ5V2XQEP3A48VE02",
		"GroupChat-01H42K4ABWQ5V2XQEP3A48VE03",
	} {
		require.NoError(t, store.Record(ctx, RecoveryTask{
			StoreID: filter.StoreID, ProjectionName: filter.ProjectionName,
			ProjectionVersion: filter.ProjectionVersion, AggregateID: aid, SeqNr: 1,
			FailureType: FailureKindEventPayload, ErrorMessage: "invalid payload",
		}))
	}
	require.NoError(t, store.Record(ctx, RecoveryTask{
		StoreID: "another-store", ProjectionName: ProjectionName, ProjectionVersion: ProjectionVersion,
		AggregateID: "GroupChat-01H42K4ABWQ5V2XQEP3A48VE04", SeqNr: 1,
		FailureType: FailureKindEventPayload, ErrorMessage: "invalid payload",
	}))

	count, err := store.CountPending(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, 3, count)

	limited, err := store.ListPending(ctx, RecoveryTaskFilter{
		StoreID: filter.StoreID, ProjectionName: filter.ProjectionName,
		ProjectionVersion: filter.ProjectionVersion, Limit: 2,
	})
	require.NoError(t, err)
	require.Len(t, limited, 2)
	for _, task := range limited {
		require.Equal(t, "store-1", task.StoreID)
	}
}

func TestHandleDynamoDBEvent_ReportsFailedAndBlockedRecords(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	broken := newReadModelFixture(t)
	healthy := newReadModelFixture(t)
	corrupt := corruptHeadEvent(2)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		broken.aid():  journalEvents(broken.aid(), broken.create, corrupt, broken.addMember),
		healthy.aid(): journalEvents(healthy.aid(), healthy.create),
	}}
	store := &fakeRecoveryTaskStore{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal,
		WithApplier(applier), WithRecoveryService(recovery))

	goodA := headRecord(t, "INSERT", broken.aid(), broken.create)
	goodB := headRecord(t, "INSERT", healthy.aid(), healthy.create)
	failingA := headRecord(t, "MODIFY", broken.aid(), corrupt)
	blockedA := headRecord(t, "MODIFY", broken.aid(), broken.addMember)
	response, err := updater.HandleDynamoDBEvent(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{goodA, goodB, failingA, blockedA},
	})
	require.NoError(t, err)
	require.Len(t, response.BatchItemFailures, 2)
	require.Equal(t, failingA.EventID, response.BatchItemFailures[0].ItemIdentifier)
	require.Equal(t, blockedA.EventID, response.BatchItemFailures[1].ItemIdentifier)

	// The healthy aggregate advanced; the broken one stopped before seq_nr 2.
	require.Equal(t, int64(1), readModelPosition(t, db, healthy.aid()))
	require.Equal(t, int64(1), readModelPosition(t, db, broken.aid()))
	require.Len(t, store.tasks, 1)
}

func TestHandleDynamoDBEvent_RecordsTaskForMalformedHeadRecord(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{}
	store := &fakeRecoveryTaskStore{}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal,
		WithApplier(applier), WithRecoveryService(recovery))

	record := dynamodbevents.DynamoDBEventRecord{
		EventID:   "malformed-1",
		EventName: "MODIFY",
		Change: dynamodbevents.DynamoDBStreamRecord{
			NewImage: map[string]dynamodbevents.DynamoDBAttributeValue{
				"aid":    dynamodbevents.NewStringAttribute(fixture.aid()),
				"seq_nr": dynamodbevents.NewNumberAttribute("2"),
				"events": dynamodbevents.NewStringAttribute("not-a-list"),
			},
		},
	}
	response, err := updater.HandleDynamoDBEvent(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{record},
	})
	require.NoError(t, err)
	require.Len(t, response.BatchItemFailures, 1)
	require.Len(t, store.tasks, 1)
	require.Equal(t, fixture.aid(), store.tasks[0].AggregateID)
	require.Equal(t, int64(2), store.tasks[0].SeqNr)
	require.Equal(t, FailureKindEventPayload, store.tasks[0].FailureType)
}

func TestHandleDynamoDBEvent_StopsAfterTransientFailure(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	broken := newReadModelFixture(t)
	healthyB := newReadModelFixture(t)
	healthyC := newReadModelFixture(t)
	journal := &fakeJournalReader{
		events: map[string][]JournalEvent{
			broken.aid(): journalEvents(broken.aid(), broken.create),
		},
		errs: []error{errors.New("connection refused")},
	}
	updater := NewReadModelUpdater(NewGroupChatDaoImpl(db), journal)

	response, err := updater.HandleDynamoDBEvent(ctx, dynamodbevents.DynamoDBEvent{
		Records: []dynamodbevents.DynamoDBEventRecord{
			headRecord(t, "MODIFY", broken.aid(), broken.addMember),
			headRecord(t, "INSERT", healthyB.aid(), healthyB.create),
			headRecord(t, "INSERT", healthyC.aid(), healthyC.create),
		},
	})
	require.NoError(t, err)
	// Once the storage is failing, the remaining records are handed back
	// instead of each spending its own retry budget.
	require.Len(t, response.BatchItemFailures, 3)
	require.Equal(t, 0, countRows(t, db, `SELECT COUNT(*) FROM read_model_positions`))
}

func TestRecoveryService_RecoverPendingSkipsOtherProjectionsAndStores(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	fixture := newReadModelFixture(t)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		fixture.aid(): journalEvents(fixture.aid(), fixture.create, fixture.addMember),
	}}
	store := &fakeRecoveryTaskStore{tasks: []RecoveryTask{
		{StoreID: "store-1", ProjectionName: ProjectionName, ProjectionVersion: ProjectionVersion,
			AggregateID: fixture.aid(), SeqNr: 2},
		{StoreID: "store-1", ProjectionName: "another-projection", ProjectionVersion: ProjectionVersion,
			AggregateID: fixture.aid(), SeqNr: 2},
		{StoreID: "another-store", ProjectionName: ProjectionName, ProjectionVersion: ProjectionVersion,
			AggregateID: fixture.aid(), SeqNr: 2},
	}}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)

	result, err := recovery.RecoverPending(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, result.Completed)
	require.Equal(t, 0, result.Pending)
	require.False(t, result.Deferred)
	require.Equal(t, int64(2), readModelPosition(t, db, fixture.aid()))
	require.Len(t, store.completed, 1)
}

func TestResync_TransientScanErrorIsRetryable(t *testing.T) {
	scanner := &fakeHeadScanner{err: errors.New("connection reset")}
	resyncer := NewResyncer(nil, scanner, nil)

	_, err := resyncer.Resync(context.Background())
	require.Error(t, err)
	var resyncErr *ResyncError
	require.ErrorAs(t, err, &resyncErr)
	require.Error(t, resyncErr.Scan)
	require.True(t, IsRetryable(err))

	permanentOnly := &ResyncError{Permanent: []error{
		permanent(FailureKindEventPayload, errors.New("bad payload")),
	}}
	require.False(t, IsRetryable(permanentOnly))
	_, isPermanent := ClassifyFailure(permanentOnly)
	require.False(t, isPermanent)
}

func TestResync_RecoveryStoreFailureAbortsScanAndIsRetryable(t *testing.T) {
	ctx := context.Background()
	db := newReadModelTestDB(t)
	broken := newReadModelFixture(t)
	healthy := newReadModelFixture(t)
	corrupt := corruptHeadEvent(2)
	journal := &fakeJournalReader{events: map[string][]JournalEvent{
		broken.aid():  journalEvents(broken.aid(), broken.create, corrupt),
		healthy.aid(): journalEvents(healthy.aid(), healthy.create, healthy.addMember),
	}}
	store := &fakeRecoveryTaskStore{recordErr: errors.New("recovery table is unavailable")}
	applier := NewApplier(NewGroupChatDaoImpl(db), journal, 1)
	recovery := NewRecoveryService(store, staticStoreID{"store-1"}, applier)
	scanner := &fakeHeadScanner{heads: []Head{
		{AggregateID: broken.aid(), TypeName: models.GroupChatIdPrefix, SeqNr: 2},
		{AggregateID: healthy.aid(), TypeName: models.GroupChatIdPrefix, SeqNr: 2},
	}}
	resyncer := NewResyncer(applier, scanner, recovery)

	_, err := resyncer.Resync(ctx)
	require.Error(t, err)
	var resyncErr *ResyncError
	require.ErrorAs(t, err, &resyncErr)
	require.Error(t, resyncErr.Scan)
	require.True(t, IsRetryable(err))
	// The scan stopped before the healthy aggregate; the next retry will get it.
	require.Equal(t, 0, countRows(t, db,
		`SELECT COUNT(*) FROM read_model_positions WHERE aggregate_id = ?`, healthy.aid()))
}

func corruptHeadEvent(seqNr int64) headEvent {
	return headEvent{
		seqNr:      seqNr,
		occurredAt: time.Unix(1700000000, 0).Add(time.Duration(seqNr) * time.Second),
		manifest:   "Broken@1",
		payload:    []byte("{not-json"),
	}
}

type staticStoreID struct {
	value string
}

func (s staticStoreID) StoreID(context.Context) (string, error) {
	return s.value, nil
}

type fakeRecoveryTaskStore struct {
	tasks     []RecoveryTask
	recordErr error
	completed []string
	listErr   error
}

func (s *fakeRecoveryTaskStore) Record(_ context.Context, task RecoveryTask) error {
	if s.recordErr != nil {
		return s.recordErr
	}
	s.tasks = append(s.tasks, task)
	return nil
}

func (s *fakeRecoveryTaskStore) ListPending(_ context.Context, filter RecoveryTaskFilter) ([]RecoveryTask, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	matched := s.matching(filter)
	if filter.Limit > 0 && len(matched) > filter.Limit {
		matched = matched[:filter.Limit]
	}
	return matched, nil
}

func (s *fakeRecoveryTaskStore) CountPending(_ context.Context, filter RecoveryTaskFilter) (int, error) {
	if s.listErr != nil {
		return 0, s.listErr
	}
	pending := 0
	for _, task := range s.matching(filter) {
		if !s.isCompleted(task) {
			pending++
		}
	}
	return pending, nil
}

func (s *fakeRecoveryTaskStore) MarkCompleted(_ context.Context, storeID, projectionName, aggregateID string, seqNr int64) error {
	s.completed = append(s.completed, recoveryTaskKey(RecoveryTask{
		StoreID: storeID, ProjectionName: projectionName, AggregateID: aggregateID, SeqNr: seqNr,
	}))
	return nil
}

func (s *fakeRecoveryTaskStore) matching(filter RecoveryTaskFilter) []RecoveryTask {
	var matched []RecoveryTask
	for _, task := range s.tasks {
		if filter.StoreID != "" && task.StoreID != filter.StoreID {
			continue
		}
		if filter.ProjectionName != "" && task.ProjectionName != filter.ProjectionName {
			continue
		}
		if filter.ProjectionVersion != "" && task.ProjectionVersion != filter.ProjectionVersion {
			continue
		}
		matched = append(matched, task)
	}
	return matched
}

func (s *fakeRecoveryTaskStore) isCompleted(task RecoveryTask) bool {
	key := recoveryTaskKey(task)
	for _, completed := range s.completed {
		if completed == key {
			return true
		}
	}
	return false
}

func recoveryTaskKey(task RecoveryTask) string {
	return fmt.Sprintf("%s/%s/%s/%d", task.StoreID, task.ProjectionName, task.AggregateID, task.SeqNr)
}

type fakeHeadScanner struct {
	heads []Head
	err   error
}

func (s *fakeHeadScanner) ScanHeads(_ context.Context, fn func(Head) error) error {
	for _, head := range s.heads {
		if err := fn(head); err != nil {
			return err
		}
	}
	return s.err
}

type fakeSleeper struct {
	delays []time.Duration
}

func (s *fakeSleeper) Sleep(_ context.Context, delay time.Duration) error {
	s.delays = append(s.delays, delay)
	return nil
}
