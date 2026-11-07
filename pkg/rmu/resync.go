package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"errors"
	"fmt"
	"log/slog"
)

// ResyncResult counts what one full reconciliation saw and did.
type ResyncResult struct {
	Heads    int
	Skipped  int
	Advanced int
	Failed   int
}

// Resyncer implements DY-15's reconciliation: scan every head, compare its
// seq_nr with the projection position, and fill the lag from the journal with
// the same position guard as the stream path.
type Resyncer struct {
	applier  *Applier
	scanner  HeadScanner
	recovery *RecoveryService
	typeName string
}

func NewResyncer(applier *Applier, scanner HeadScanner, recovery *RecoveryService) *Resyncer {
	return &Resyncer{
		applier:  applier,
		scanner:  scanner,
		recovery: recovery,
		typeName: models.GroupChatIdPrefix,
	}
}

// Resync reads the whole head table with strong consistency. A permanent
// failure is recorded as a recovery task and does not stop the remaining
// aggregates; a storage failure aborts the scan so the caller can retry it.
func (r *Resyncer) Resync(ctx context.Context) (ResyncResult, error) {
	var result ResyncResult
	var errs []error
	scanErr := r.scanner.ScanHeads(ctx, func(head Head) error {
		if head.AggregateID == configAggregateID || head.TypeName != r.typeName {
			result.Skipped++
			return nil
		}
		result.Heads++
		position, err := r.applier.Position(ctx, head.AggregateID)
		if err != nil {
			return fmt.Errorf("read position for %s: %w", head.AggregateID, err)
		}
		if head.SeqNr <= position {
			return nil
		}
		result.Advanced++
		if err := r.applier.ApplyThrough(ctx, head.AggregateID, head.SeqNr); err != nil {
			if _, ok := ClassifyFailure(err); !ok {
				return fmt.Errorf("resync %s to seq_nr %d: %w", head.AggregateID, head.SeqNr, err)
			}
			result.Failed++
			if r.recovery != nil {
				if recordErr := r.recovery.Record(ctx, head.AggregateID, head.SeqNr, err); recordErr != nil {
					errs = append(errs, recordErr)
					return nil
				}
			}
			errs = append(errs, fmt.Errorf("resync %s to seq_nr %d: %w", head.AggregateID, head.SeqNr, err))
			slog.Error("resync could not apply aggregate; recorded a recovery task",
				"aid", head.AggregateID, "seq_nr", head.SeqNr, "error", err)
			return nil
		}
		return nil
	})
	if scanErr != nil {
		return result, errors.Join(append(errs, scanErr)...)
	}
	return result, errors.Join(errs...)
}
