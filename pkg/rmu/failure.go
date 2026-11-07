package rmu

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/aws/smithy-go"
	"github.com/go-sql-driver/mysql"
)

// FailureKind is the reason a record could not be applied. It is stored with a
// recovery task so an operator can tell a data problem from a storage outage.
type FailureKind string

const (
	FailureKindEventPayload        FailureKind = "event_payload"
	FailureKindUnsupportedEvent    FailureKind = "unsupported_event"
	FailureKindJournalGap          FailureKind = "journal_gap"
	FailureKindReadModelConstraint FailureKind = "read_model_constraint"
	FailureKindUnknown             FailureKind = "unknown"
)

// PermanentFailure marks an error that retrying the same record cannot fix.
// Only payload, data-gap, and constraint problems are permanent; storage
// outages stay unclassified so the caller retries them.
type PermanentFailure struct {
	Kind FailureKind
	Err  error
}

func (e *PermanentFailure) Error() string {
	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

func (e *PermanentFailure) Unwrap() error {
	return e.Err
}

func permanent(kind FailureKind, err error) error {
	if err == nil {
		return nil
	}
	return &PermanentFailure{Kind: kind, Err: err}
}

// RecoveryRecordError means the event failed and the recovery task could not be
// written. It deliberately unwraps only the write failure, so callers treat it
// as retryable instead of skipping a record whose task was lost.
type RecoveryRecordError struct {
	ApplyErr  error
	RecordErr error
}

func (e *RecoveryRecordError) Error() string {
	return fmt.Sprintf("record recovery task after apply failure (%v): %v", e.ApplyErr, e.RecordErr)
}

func (e *RecoveryRecordError) Unwrap() error {
	return e.RecordErr
}

// ClassifyFailure reports whether the failure is permanent and which kind of
// recovery task it should create. A false return means the error is retryable.
func ClassifyFailure(err error) (FailureKind, bool) {
	if err == nil {
		return "", false
	}
	var recoveryRecordErr *RecoveryRecordError
	if errors.As(err, &recoveryRecordErr) {
		// The task was not written. The caller must retry rather than skip a
		// record whose recovery task would be lost.
		return FailureKindUnknown, false
	}
	var resyncErr *ResyncError
	if errors.As(err, &resyncErr) {
		// A resync reports permanent per-aggregate failures alongside a scan
		// error. Whether to retry depends on the scan, which IsRetryable owns.
		return FailureKindUnknown, false
	}
	var permanentFailure *PermanentFailure
	if errors.As(err, &permanentFailure) {
		return permanentFailure.Kind, true
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062, 1452, 1406, 1264, 1366:
			return FailureKindReadModelConstraint, true
		}
	}
	return FailureKindUnknown, false
}

// IsRetryable reports whether the same operation may succeed later. Known data
// problems are not retried. Database, network, and throttling failures are.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	var resyncErr *ResyncError
	if errors.As(err, &resyncErr) {
		return resyncErr.Scan != nil && IsRetryable(resyncErr.Scan)
	}
	if _, ok := ClassifyFailure(err); ok {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1040, 1053, 1205, 1213, 2002, 2003, 2006, 2013:
			return true
		}
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ThrottlingException", "Throttling", "ProvisionedThroughputExceededException",
			"RequestLimitExceeded", "InternalServerError", "ServiceUnavailable",
			"ExpiredIteratorException", "TransactionConflictException":
			return true
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return true
	}
	// Unknown storage errors are retried a bounded number of times. A record
	// that still fails is reported to Lambda and retried or sent to the DLQ.
	return true
}
