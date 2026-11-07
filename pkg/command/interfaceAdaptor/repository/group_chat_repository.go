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

type GroupChatRepository interface {
	Store(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error]
	StoreEvent(event events.GroupChatEvent, version uint64) mo.Option[error]
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
		es.snapshotDecider = func(event events.GroupChatEvent, _ *domain.GroupChat) bool {
			return event.GetSeqNr()%numberOfEvens == 0
		}
		return nil
	}
}

func (g *GroupChatRepositoryImpl) Store(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error] {
	if event.IsCreated() || g.snapshotDecider != nil && g.snapshotDecider(event, snapshot) {
		return g.StoreEventWithSnapshot(event, snapshot)
	} else {
		return g.StoreEvent(event, snapshot.GetVersion())
	}
}

// StoreEvent appends only the event. The expected previous head is derived from
// the event's sequence number, so version is no longer used by the store.
func (g *GroupChatRepositoryImpl) StoreEvent(event events.GroupChatEvent, _ uint64) mo.Option[error] {
	envelope, err := newEventEnvelope(event)
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
	envelope, err := newEventEnvelope(event)
	if err != nil {
		return mo.Some(err)
	}
	snapshotEnvelope, err := esa.NewSnapshotEnvelope(*snapshot, esa.SeqNr(event.GetSeqNr()))
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
	payloads := make([]events.GroupChatEvent, 0, len(eventsByIdSinceSeqNr))
	for _, envelope := range eventsByIdSinceSeqNr {
		payloads = append(payloads, envelope.Payload())
	}
	return mo.Ok[mo.Option[domain.GroupChat]](mo.Some(domain.ReplayGroupChat(payloads, snapshot)))
}

// newEventEnvelope bridges the domain event to the v5 envelope API. Metadata
// still lives in the payload until the envelope migration in #667-#670.
func newEventEnvelope(event events.GroupChatEvent) (esa.EventEnvelope[events.GroupChatEvent], error) {
	occurredAt := time.Unix(0, int64(event.GetOccurredAt())*int64(time.Millisecond))
	return esa.NewEventEnvelope[events.GroupChatEvent](
		event.GetAggregateId(),
		esa.SeqNr(event.GetSeqNr()),
		occurredAt,
		event,
	)
}
