package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/samber/mo"
)

type GroupChatRepository interface {
	StoreEvent(event events.GroupChatEvent, version uint64) mo.Option[error]
	StoreEventWithSnapshot(event events.GroupChatEvent, snapshot *domain.GroupChat) mo.Option[error]
	FindById(id *models.GroupChatId) mo.Result[domain.GroupChat]
}

type GroupChatRepositoryImpl struct {
	eventStore esa.EventStore
}

func NewGroupChatRepository(eventStore esa.EventStore) GroupChatRepositoryImpl {
	return GroupChatRepositoryImpl{eventStore}
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

func (g *GroupChatRepositoryImpl) FindById(id *models.GroupChatId) mo.Result[domain.GroupChat] {
	result, err := g.eventStore.GetLatestSnapshotById(id)
	if err != nil {
		return mo.Err[domain.GroupChat](err)
	}
	if result.Empty() {
		return mo.Err[domain.GroupChat](fmt.Errorf("not found"))
	} else {
		eventsByIdSinceSeqNr, err := g.eventStore.GetEventsByIdSinceSeqNr(id, result.Aggregate().GetSeqNr()+1)
		if err != nil {
			return mo.Err[domain.GroupChat](err)
		}
		return mo.Ok[domain.GroupChat](domain.ReplayGroupChat(eventsByIdSinceSeqNr, *result.Aggregate().(*domain.GroupChat)))
	}
}
