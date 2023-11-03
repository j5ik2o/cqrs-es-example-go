package events

import (
	"cqrs-es-example-go/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"time"
)

type GroupChatDeleted struct {
	id          string
	aggregateId models.GroupChatId
	seqNr       uint64
	occurredAt  uint64
}

func NewGroupChatDeleted(id string, aggregateId models.GroupChatId, seqNr uint64) *GroupChatDeleted {
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatDeleted{id, aggregateId, seqNr, occurredAt}
}

func (g *GroupChatDeleted) GetId() string {
	return g.id
}

func (g *GroupChatDeleted) GetTypeName() string {
	return "group-chat-deleted"
}

func (g *GroupChatDeleted) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatDeleted) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatDeleted) IsCreated() bool {
	return true
}

func (g *GroupChatDeleted) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatDeleted) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.seqNr, g.occurredAt)
}
