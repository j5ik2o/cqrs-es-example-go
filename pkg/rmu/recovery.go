package rmu

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

const (
	// DefaultRecoveryTaskLimit bounds one recovery pass so it fits inside the
	// Lambda timeout. Tasks beyond the bound are deferred to the next run.
	DefaultRecoveryTaskLimit = 100
	// MaxRecoveryTaskLimit is the largest accepted bound. A bigger value is
	// clamped, not rejected, so a bad environment variable cannot stall the
	// worker forever.
	MaxRecoveryTaskLimit = 1000
)

// NormalizeRecoveryTaskLimit keeps the configured bound inside the range one
// pass can finish. Zero, negative, and oversized values fall back to the
// nearest bound.
func NormalizeRecoveryTaskLimit(limit int) int {
	if limit <= 0 {
		return DefaultRecoveryTaskLimit
	}
	if limit > MaxRecoveryTaskLimit {
		return MaxRecoveryTaskLimit
	}
	return limit
}

// RecoveryPendingResult summarises one pending-task replay. Pending is the
// count left in the table after the pass, so it includes tasks that failed.
type RecoveryPendingResult struct {
	Completed int
	Failed    int
	Pending   int
	Deferred  bool
	Limit     int
}

// RecoveryService records failed events and replays them from the journal. It is
// the DLQ half of the RMU contract: the task is durable, the replay is
// idempotent, and completion is claimed only after the position reaches the
// task's seq_nr. SQS is a delivery detail outside this package; a message only
// wakes a worker, which re-reads the event body from the journal.
type RecoveryService struct {
	store             RecoveryTaskStore
	storeID           StoreIDProvider
	applier           *Applier
	projectionName    string
	projectionVersion string
	taskLimit         int
}

type RecoveryOption func(*RecoveryService)

func WithProjection(name, version string) RecoveryOption {
	return func(s *RecoveryService) {
		s.projectionName = name
		s.projectionVersion = version
	}
}

// WithRecoveryTaskLimit bounds how many pending tasks one pass processes. The
// value is normalised, so an invalid setting cannot remove the bound.
func WithRecoveryTaskLimit(limit int) RecoveryOption {
	return func(s *RecoveryService) {
		s.taskLimit = NormalizeRecoveryTaskLimit(limit)
	}
}

func NewRecoveryService(store RecoveryTaskStore, storeID StoreIDProvider, applier *Applier, opts ...RecoveryOption) *RecoveryService {
	service := &RecoveryService{
		store:             store,
		storeID:           storeID,
		applier:           applier,
		projectionName:    ProjectionName,
		projectionVersion: ProjectionVersion,
		taskLimit:         DefaultRecoveryTaskLimit,
	}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

// Record persists a recovery task. A failure here is returned to the caller on
// purpose: losing the task would hide a permanently failing event, so the
// record must fail and be retried by the delivery mechanism.
func (s *RecoveryService) Record(ctx context.Context, aggregateID string, seqNr int64, failure error) error {
	if s == nil {
		return nil
	}
	kind, _ := ClassifyFailure(failure)
	if kind == "" {
		kind = FailureKindUnknown
	}
	storeID, err := s.storeID.StoreID(ctx)
	if err != nil {
		return fmt.Errorf("read store id for recovery task: %w", err)
	}
	task := RecoveryTask{
		StoreID:           storeID,
		ProjectionName:    s.projectionName,
		ProjectionVersion: s.projectionVersion,
		AggregateID:       aggregateID,
		SeqNr:             seqNr,
		FailureType:       kind,
		ErrorMessage:      failure.Error(),
		Status:            RecoveryStatusPending,
	}
	if err := s.store.Record(ctx, task); err != nil {
		return fmt.Errorf("record recovery task for %s seq_nr %d: %w", aggregateID, seqNr, err)
	}
	return nil
}

// Recover replays one task. It re-reads the position first, so calling it twice
// is safe: the second call finds the position at or beyond the task and only
// marks the task complete.
func (s *RecoveryService) Recover(ctx context.Context, task RecoveryTask) error {
	if s == nil {
		return nil
	}
	if task.ProjectionName != s.projectionName || task.ProjectionVersion != s.projectionVersion {
		return fmt.Errorf("recovery task targets projection %s/%s, not %s/%s",
			task.ProjectionName, task.ProjectionVersion, s.projectionName, s.projectionVersion)
	}
	storeID, err := s.storeID.StoreID(ctx)
	if err != nil {
		return fmt.Errorf("read store id for recovery task: %w", err)
	}
	if task.StoreID != storeID {
		return fmt.Errorf("recovery task belongs to store %s, not %s", task.StoreID, storeID)
	}
	position, err := s.applier.Position(ctx, task.AggregateID)
	if err != nil {
		return fmt.Errorf("read position for %s: %w", task.AggregateID, err)
	}
	if position < task.SeqNr {
		if err := s.applier.ApplyThrough(ctx, task.AggregateID, task.SeqNr); err != nil {
			return fmt.Errorf("recover %s seq_nr %d: %w", task.AggregateID, task.SeqNr, err)
		}
	}
	position, err = s.applier.Position(ctx, task.AggregateID)
	if err != nil {
		return fmt.Errorf("verify position for %s: %w", task.AggregateID, err)
	}
	if position < task.SeqNr {
		return fmt.Errorf("aggregate %s stayed at position %d, expected %d", task.AggregateID, position, task.SeqNr)
	}
	if err := s.store.MarkCompleted(ctx, task.StoreID, task.ProjectionName, task.AggregateID, task.SeqNr); err != nil {
		return fmt.Errorf("mark recovery task complete for %s seq_nr %d: %w", task.AggregateID, task.SeqNr, err)
	}
	return nil
}

// RecoverPending replays pending tasks in creation order and returns the tasks
// that still fail. One failing aggregate does not stop the others. The pass
// reads at most one task beyond the limit to detect leftovers; the extra task
// and everything after it stay pending for the next run.
func (s *RecoveryService) RecoverPending(ctx context.Context) (RecoveryPendingResult, error) {
	if s == nil {
		return RecoveryPendingResult{}, nil
	}
	storeID, err := s.storeID.StoreID(ctx)
	if err != nil {
		return RecoveryPendingResult{}, fmt.Errorf("read store id for recovery tasks: %w", err)
	}
	limit := NormalizeRecoveryTaskLimit(s.taskLimit)
	filter := RecoveryTaskFilter{
		StoreID:           storeID,
		ProjectionName:    s.projectionName,
		ProjectionVersion: s.projectionVersion,
	}
	tasks, err := s.store.ListPending(ctx, RecoveryTaskFilter{
		StoreID:           filter.StoreID,
		ProjectionName:    filter.ProjectionName,
		ProjectionVersion: filter.ProjectionVersion,
		Limit:             limit + 1,
	})
	if err != nil {
		return RecoveryPendingResult{}, fmt.Errorf("list recovery tasks: %w", err)
	}
	result := RecoveryPendingResult{Limit: limit, Deferred: len(tasks) > limit}
	if result.Deferred {
		tasks = tasks[:limit]
	}
	var errs []error
	for _, task := range tasks {
		if err := s.Recover(ctx, task); err != nil {
			result.Failed++
			errs = append(errs, err)
			continue
		}
		result.Completed++
	}
	pending, countErr := s.store.CountPending(ctx, filter)
	if countErr != nil {
		errs = append(errs, fmt.Errorf("count pending recovery tasks: %w", countErr))
	}
	result.Pending = pending
	if result.Deferred {
		slog.Warn("recovery task limit reached; remaining tasks are deferred to the next run",
			"limit", limit, "completed", result.Completed, "pending", pending)
	}
	return result, errors.Join(errs...)
}
