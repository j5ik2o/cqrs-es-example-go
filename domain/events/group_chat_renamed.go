package events

import (
	"cqrs-es-example-go/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

type GroupChatRenamed struct {
	id          string
	aggregateId *models.GroupChatId
	name        *models.GroupChatName
	seqNr       uint64
	executorId  *models.UserAccountId
	occurredAt  uint64
}

func NewGroupChatRenamed(aggregateId *models.GroupChatId, seqNr uint64, name *models.GroupChatName, executorId *models.UserAccountId) *GroupChatRenamed {
	id := ulid.Make()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatRenamed{id: id.String(), aggregateId: aggregateId, name: name, seqNr: seqNr, executorId: executorId, occurredAt: occurredAt}
}

func (g *GroupChatRenamed) GetId() string {
	return g.id
}

func (g *GroupChatRenamed) GetTypeName() string {
	return "group-chat-renamed"
}

func (g *GroupChatRenamed) GetAggregateId() esa.AggregateId {
	return g.aggregateId
}

func (g *GroupChatRenamed) GetName() *models.GroupChatName {
	return g.name
}

func (g *GroupChatRenamed) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatRenamed) GetExecutorId() *models.UserAccountId {
	return g.executorId
}

func (g *GroupChatRenamed) IsCreated() bool {
	return false
}

func (g *GroupChatRenamed) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatRenamed) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s name: %s, seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.name, g.seqNr, g.occurredAt)
}
