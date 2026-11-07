package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"time"
)

// Applier owns the single position rule: only the event at position+1 is
// applied, and the read model change and the position advance commit together.
// Both the stream path and the journal replay path (resync, recovery) use it.
type Applier struct {
	dao                     GroupChatDao
	journal                 JournalReader
	maxEventsPerTransaction int
	readAheadEvents         int
}

func NewApplier(dao GroupChatDao, journal JournalReader, maxEventsPerTransaction int) *Applier {
	if maxEventsPerTransaction < 1 {
		maxEventsPerTransaction = 1
	}
	return &Applier{
		dao:                     dao,
		journal:                 journal,
		maxEventsPerTransaction: maxEventsPerTransaction,
		readAheadEvents:         100,
	}
}

// Position returns the last applied seq_nr for the aggregate.
func (a *Applier) Position(ctx context.Context, aggregateID string) (int64, error) {
	return a.dao.GetPosition(ctx, aggregateID)
}

// applyHeadEvent applies a head record. The journal fills a gap below the
// record's seq_nr; the record itself carries the newest event.
func (a *Applier) applyHeadEvent(ctx context.Context, aggregateID string, event headEvent) error {
	groupChatID, err := models.NewGroupChatIdFromString(aggregateID).Get()
	if err != nil {
		return permanent(FailureKindEventPayload, err)
	}
	currentSeqNr, err := a.dao.GetPosition(ctx, aggregateID)
	if err != nil {
		return err
	}
	if event.seqNr <= currentSeqNr {
		return nil
	}

	pending := []headEvent{event}
	if event.seqNr > currentSeqNr+1 {
		if a.journal == nil {
			return permanent(FailureKindJournalGap,
				fmt.Errorf("aggregate %s: seq_nr %d follows position %d and no journal reader is configured", aggregateID, event.seqNr, currentSeqNr))
		}
		journalEvents, err := a.journal.ReadEvents(ctx, aggregateID, currentSeqNr+1, event.seqNr-1)
		if err != nil {
			return fmt.Errorf("aggregate %s: read journal seq_nr %d..%d: %w", aggregateID, currentSeqNr+1, event.seqNr-1, err)
		}
		if err := requireContiguous(currentSeqNr+1, event.seqNr-1, journalEvents); err != nil {
			return permanent(FailureKindJournalGap, fmt.Errorf("aggregate %s: %w", aggregateID, err))
		}
		gap := make([]headEvent, len(journalEvents))
		for i, journalEvent := range journalEvents {
			gap[i] = headEventFromJournal(journalEvent)
		}
		pending = append(gap, event)
	}
	return a.applyEvents(ctx, aggregateID, &groupChatID, pending)
}

// ApplyThrough replays the journal from the current position up to targetSeqNr.
// It is the shared body of resync and DLQ recovery, and it never uses the head
// payload, so a repaired journal entry is picked up on the next run.
func (a *Applier) ApplyThrough(ctx context.Context, aggregateID string, targetSeqNr int64) error {
	if targetSeqNr <= 0 {
		return nil
	}
	groupChatID, err := models.NewGroupChatIdFromString(aggregateID).Get()
	if err != nil {
		return permanent(FailureKindEventPayload, err)
	}
	if a.journal == nil {
		return permanent(FailureKindJournalGap,
			fmt.Errorf("aggregate %s: journal reader is required to reach seq_nr %d", aggregateID, targetSeqNr))
	}
	for {
		currentSeqNr, err := a.dao.GetPosition(ctx, aggregateID)
		if err != nil {
			return err
		}
		if currentSeqNr >= targetSeqNr {
			return nil
		}
		endSeqNr := currentSeqNr + int64(a.readAheadEvents)
		if endSeqNr > targetSeqNr {
			endSeqNr = targetSeqNr
		}
		journalEvents, err := a.journal.ReadEvents(ctx, aggregateID, currentSeqNr+1, endSeqNr)
		if err != nil {
			return fmt.Errorf("aggregate %s: read journal seq_nr %d..%d: %w", aggregateID, currentSeqNr+1, endSeqNr, err)
		}
		if err := requireContiguous(currentSeqNr+1, endSeqNr, journalEvents); err != nil {
			return permanent(FailureKindJournalGap, fmt.Errorf("aggregate %s: %w", aggregateID, err))
		}
		events := make([]headEvent, len(journalEvents))
		for i, journalEvent := range journalEvents {
			events[i] = headEventFromJournal(journalEvent)
		}
		if err := a.applyEvents(ctx, aggregateID, &groupChatID, events); err != nil {
			return err
		}
		afterSeqNr, err := a.dao.GetPosition(ctx, aggregateID)
		if err != nil {
			return err
		}
		if afterSeqNr <= currentSeqNr {
			return fmt.Errorf("aggregate %s: position stayed at %d while replaying to %d", aggregateID, afterSeqNr, targetSeqNr)
		}
	}
}

func (a *Applier) applyEvents(ctx context.Context, aggregateID string, groupChatID *models.GroupChatId, events []headEvent) error {
	for len(events) > 0 {
		batchSize := a.maxEventsPerTransaction
		if batchSize > len(events) {
			batchSize = len(events)
		}
		batch := events[:batchSize]
		if err := a.dao.Process(ctx, aggregateID, func(w ReadModelWriter, currentSeqNr int64) (int64, error) {
			nextSeqNr := currentSeqNr
			for _, event := range batch {
				if event.seqNr <= currentSeqNr {
					continue
				}
				if event.seqNr != nextSeqNr+1 {
					return nextSeqNr, permanent(FailureKindJournalGap,
						fmt.Errorf("expected seq_nr %d but got %d", nextSeqNr+1, event.seqNr))
				}
				if err := applyGroupChatEvent(groupChatID, event, w); err != nil {
					return nextSeqNr, fmt.Errorf("apply seq_nr %d: %w", event.seqNr, err)
				}
				nextSeqNr = event.seqNr
			}
			return nextSeqNr, nil
		}); err != nil {
			return err
		}
		events = events[batchSize:]
	}
	return nil
}

func headEventFromJournal(journalEvent JournalEvent) headEvent {
	return headEvent{
		seqNr:      journalEvent.SeqNr,
		occurredAt: journalEvent.OccurredAt,
		manifest:   journalEvent.Manifest,
		payload:    journalEvent.Payload,
	}
}

// headEvent is one envelope entry in the head item's events list or in the
// journal.
type headEvent struct {
	seqNr      int64
	occurredAt time.Time
	manifest   string
	payload    []byte
}
