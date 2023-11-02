package events

import (
	"cqrs-es-example-go/domain"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"time"
)

type GroupChatRenamed struct {
	id          string
	aggregateId domain.GroupChatId
	name        string
	seqNr       uint64
	occurredAt  uint64
}

func NewGroupChatRenamed(id string, aggregateId domain.GroupChatId, name string, seqNr uint64) *GroupChatRenamed {
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatRenamed{id, aggregateId, name, seqNr, occurredAt}
}

func (g *GroupChatRenamed) GetId() string {
	return g.id
}

func (g *GroupChatRenamed) GetTypeName() string {
	return "group-chat-renamed"
}

func (g *GroupChatRenamed) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatRenamed) GetName() string {
	return g.name
}

func (g *GroupChatRenamed) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatRenamed) IsCreated() bool {
	return true
}

func (g *GroupChatRenamed) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatRenamed) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s name: %s, seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.name, g.seqNr, g.occurredAt)
}
