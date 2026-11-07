package rmu

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	RecoveryStatusPending   = "pending"
	RecoveryStatusCompleted = "completed"

	// ProjectionName and ProjectionVersion identify the read model this RMU
	// maintains. A recovery task is scoped to them so two projections over the
	// same journal do not consume each other's tasks.
	ProjectionName    = "group-chat-read-model"
	ProjectionVersion = "1"
)

// RecoveryTask is the durable record of one event that could not be applied.
// It stores identifiers only. The event body is always re-read from the
// journal, so a task cannot go stale when the payload is repaired.
type RecoveryTask struct {
	StoreID           string
	ProjectionName    string
	ProjectionVersion string
	AggregateID       string
	SeqNr             int64
	FailureType       FailureKind
	ErrorMessage      string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RecoveryTaskStore persists recovery tasks. Recording is idempotent on the
// (store, projection, aggregate, seq_nr) key so a redelivered record does not
// create duplicates.
type RecoveryTaskStore interface {
	Record(ctx context.Context, task RecoveryTask) error
	ListPending(ctx context.Context) ([]RecoveryTask, error)
	MarkCompleted(ctx context.Context, storeID, projectionName, aggregateID string, seqNr int64) error
}

type MySQLRecoveryTaskStore struct {
	db *sqlx.DB
}

func NewMySQLRecoveryTaskStore(db *sqlx.DB) *MySQLRecoveryTaskStore {
	return &MySQLRecoveryTaskStore{db: db}
}

func (s *MySQLRecoveryTaskStore) Record(ctx context.Context, task RecoveryTask) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO read_model_recovery_tasks
			(store_id, projection_name, projection_version, aggregate_id, seq_nr,
			 failure_type, error_message, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			projection_version = VALUES(projection_version),
			failure_type = VALUES(failure_type),
			error_message = VALUES(error_message),
			status = VALUES(status),
			updated_at = VALUES(updated_at)`,
		task.StoreID, task.ProjectionName, task.ProjectionVersion, task.AggregateID,
		task.SeqNr, string(task.FailureType), task.ErrorMessage, RecoveryStatusPending,
		now, now)
	return err
}

func (s *MySQLRecoveryTaskStore) ListPending(ctx context.Context) ([]RecoveryTask, error) {
	var rows []struct {
		StoreID           string    `db:"store_id"`
		ProjectionName    string    `db:"projection_name"`
		ProjectionVersion string    `db:"projection_version"`
		AggregateID       string    `db:"aggregate_id"`
		SeqNr             int64     `db:"seq_nr"`
		FailureType       string    `db:"failure_type"`
		ErrorMessage      string    `db:"error_message"`
		Status            string    `db:"status"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}
	if err := s.db.SelectContext(ctx, &rows, `
		SELECT store_id, projection_name, projection_version, aggregate_id, seq_nr,
		       failure_type, error_message, status, created_at, updated_at
		FROM read_model_recovery_tasks
		WHERE status = ?
		ORDER BY created_at, aggregate_id, seq_nr`, RecoveryStatusPending); err != nil {
		return nil, err
	}
	tasks := make([]RecoveryTask, len(rows))
	for i, row := range rows {
		tasks[i] = RecoveryTask{
			StoreID:           row.StoreID,
			ProjectionName:    row.ProjectionName,
			ProjectionVersion: row.ProjectionVersion,
			AggregateID:       row.AggregateID,
			SeqNr:             row.SeqNr,
			FailureType:       FailureKind(row.FailureType),
			ErrorMessage:      row.ErrorMessage,
			Status:            row.Status,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
		}
	}
	return tasks, nil
}

func (s *MySQLRecoveryTaskStore) MarkCompleted(ctx context.Context, storeID, projectionName, aggregateID string, seqNr int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE read_model_recovery_tasks
		SET status = ?, updated_at = ?
		WHERE store_id = ? AND projection_name = ? AND aggregate_id = ? AND seq_nr = ?`,
		RecoveryStatusCompleted, time.Now(), storeID, projectionName, aggregateID, seqNr)
	return err
}
