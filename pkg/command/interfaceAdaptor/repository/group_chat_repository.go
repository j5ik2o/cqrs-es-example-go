package repository

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"time"

	esa "github.com/j5ik2o/event-store-adapter-go/v5"
	"github.com/samber/mo"
)

// SnapshotManifest is the envelope manifest for the GroupChat aggregate state.
const SnapshotManifest = "GroupChat@1"

type GroupChatRepository interface {
	Store(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error]
	StoreEvent(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error]
	StoreEventWithSnapshot(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error]
	FindById(id *models.GroupChatId) mo.Result[mo.Option[domain.GroupChat]]
}
type GroupChatRepositoryOption func(*GroupChatRepositoryImpl) error
type SnapshotDecider = func(event events.GroupChatEvent, snapshot *domain.GroupChat) bool

type GroupChatRepositoryImpl struct {
	eventStore      esa.EventStore[events.GroupChatEvent, domain.GroupChat]
	snapshotDecider SnapshotDecider
}

func NewGroupChatRepository(eventStore esa.EventStore[events.GroupChatEvent, domain.GroupChat], options ...GroupChatRepositoryOption) (GroupChatRepositoryImpl, error) {
	repository := GroupChatRepositoryImpl{eventStore, nil}
	for _, option := range options {
		if err := option(&repository); err != nil {
			return GroupChatRepositoryImpl{}, err
		}
	}
	return repository, nil
}

func WithSnapshotDecider(snapshotDecider SnapshotDecider) GroupChatRepositoryOption {
	return func(es *GroupChatRepositoryImpl) error {
		es.snapshotDecider = snapshotDecider
		return nil
	}
}

func WithRetention(numberOfEvens uint64) GroupChatRepositoryOption {
	return func(es *GroupChatRepositoryImpl) error {
		es.snapshotDecider = func(_ events.GroupChatEvent, snapshot *domain.GroupChat) bool {
			return snapshot.GetSeqNr()%numberOfEvens == 0
		}
		return nil
	}
}

func (g *GroupChatRepositoryImpl) Store(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error] {
	if snapshot.GetSeqNr() == 1 || g.snapshotDecider != nil && g.snapshotDecider(event, snapshot) {
		return g.StoreEventWithSnapshot(event, snapshot)
	}
	return g.StoreEvent(event, snapshot)
}

// StoreEvent appends only the event. The expected previous head is derived from
// the event's sequence number, so the store needs no version argument.
func (g *GroupChatRepositoryImpl) StoreEvent(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error] {
	envelope, err := newEventEnvelope(event, snapshot)
	if err != nil {
		return mo.Some(err)
	}
	err = g.eventStore.PersistEvent(context.Background(), envelope)
	if err != nil {
		return mo.Some(err)
	}
	return mo.None[error]()
}

func (g *GroupChatRepositoryImpl) StoreEventWithSnapshot(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error] {
	envelope, err := newEventEnvelope(event, snapshot)
	if err != nil {
		return mo.Some(err)
	}
	snapshotEnvelope, err := newSnapshotEnvelope(snapshot)
	if err != nil {
		return mo.Some(err)
	}
	err = g.eventStore.PersistEventAndSnapshot(context.Background(), envelope, snapshotEnvelope)
	if err != nil {
		return mo.Some(err)
	}
	return mo.None[error]()
}

func (g *GroupChatRepositoryImpl) FindById(id *models.GroupChatId) mo.Result[mo.Option[domain.GroupChat]] {
	read, err := g.eventStore.GetLatestSnapshotByID(context.Background(), id)
	if err != nil {
		return mo.Err[mo.Option[domain.GroupChat]](err)
	}
	if read == nil {
		return mo.Ok[mo.Option[domain.GroupChat]](mo.None[domain.GroupChat]())
	}

	snapshot := domain.GroupChat{}
	since := esa.SeqNr(1)
	if read.Snapshot != nil {
		snapshot = read.Snapshot.Aggregate()
		since = read.Snapshot.SeqNr() + 1
	}

	eventsByIdSinceSeqNr, err := g.eventStore.GetEventsByIDSinceSeqNr(context.Background(), id, since)
	if err != nil {
		return mo.Err[mo.Option[domain.GroupChat]](err)
	}
	records := make([]domain.EventRecord, 0, len(eventsByIdSinceSeqNr))
	for _, envelope := range eventsByIdSinceSeqNr {
		records = append(records, domain.EventRecord{
			Event: envelope.Payload(),
			SeqNr: uint64(envelope.SeqNr()),
		})
	}
	return mo.Ok[mo.Option[domain.GroupChat]](mo.Some(domain.ReplayGroupChat(*id, records, snapshot)))
}

// newEventEnvelope puts the event metadata in the envelope. The payload stays
// pure domain data and the manifest names the event type and schema version.
func newEventEnvelope(event events.GroupChatEvent, snapshot *domain.GroupChat) (esa.EventEnvelope[events.GroupChatEvent], error) {
	return esa.NewEventEnvelope(
		snapshot.GetGroupChatId(),
		esa.SeqNr(snapshot.GetSeqNr()),
		time.Now(),
		event,
		esa.WithManifest(events.Manifest(event.GetTypeName())),
	)
}

// newSnapshotEnvelope puts the snapshot metadata in the envelope. The seq_nr
// is held by the envelope as well as by the aggregate state payload.
func newSnapshotEnvelope(snapshot *domain.GroupChat) (esa.SnapshotEnvelope[domain.GroupChat], error) {
	return esa.NewSnapshotEnvelope(
		*snapshot,
		esa.SeqNr(snapshot.GetSeqNr()),
		esa.WithManifest(SnapshotManifest),
	)
}
