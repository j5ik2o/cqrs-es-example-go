package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
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
	eventStore      esa.EventStore
	snapshotDecider SnapshotDecider
}

func NewGroupChatRepository(eventStore esa.EventStore, options ...GroupChatRepositoryOption) (GroupChatRepositoryImpl, error) {
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

func (g *GroupChatRepositoryImpl) StoreEvent(event events.GroupChatEvent, version uint64) mo.Option[error] {
	if err := g.eventStore.PersistEvent(event, version); err != nil {
		return mo.Some(err)
	} else {
		return mo.None[error]()
	}
}

func (g *GroupChatRepositoryImpl) StoreEventWithSnapshot(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error] {
	if err := g.eventStore.PersistEventAndSnapshot(event, snapshot); err != nil {
		return mo.Some(err)
	} else {
		return mo.None[error]()
	}
}

func (g *GroupChatRepositoryImpl) FindById(id *models.GroupChatId) mo.Result[mo.Option[domain.GroupChat]] {
	result, err := g.eventStore.GetLatestSnapshotById(id)
	if err != nil {
		return mo.Err[mo.Option[domain.GroupChat]](err)
	}
	if result.Empty() {
		return mo.Ok[mo.Option[domain.GroupChat]](mo.None[domain.GroupChat]())
	} else {
		eventsByIdSinceSeqNr, err := g.eventStore.GetEventsByIdSinceSeqNr(id, result.Aggregate().GetSeqNr()+1)
		if err != nil {
			return mo.Err[mo.Option[domain.GroupChat]](err)
		}
		return mo.Ok[mo.Option[domain.GroupChat]](mo.Some(domain.ReplayGroupChat(eventsByIdSinceSeqNr, *result.Aggregate().(*domain.GroupChat))))
	}
}
