package events

import (
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

type GroupChatRenamed struct {
	id          string
	aggregateId models2.GroupChatId
	name        models2.GroupChatName
	seqNr       uint64
	executorId  models2.UserAccountId
	occurredAt  uint64
}

func NewGroupChatRenamed(aggregateId models2.GroupChatId, name models2.GroupChatName, seqNr uint64, executorId models2.UserAccountId) *GroupChatRenamed {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatRenamed{id, aggregateId, name, seqNr, executorId, occurredAt}
}

func NewGroupChatRenamedFrom(id string, aggregateId models2.GroupChatId, name models2.GroupChatName, seqNr uint64, executorId models2.UserAccountId, occurredAt uint64) *GroupChatRenamed {
	return &GroupChatRenamed{id, aggregateId, name, seqNr, executorId, occurredAt}
}

func (g *GroupChatRenamed) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":           g.id,
		"aggregate_id": g.aggregateId.ToJSON(),
		"name":         g.name.ToJSON(),
		"seq_nr":       g.seqNr,
		"executor_id":  g.executorId.ToJSON(),
		"occurred_at":  g.occurredAt,
	}
}

func (g *GroupChatRenamed) GetId() string {
	return g.id
}

func (g *GroupChatRenamed) GetTypeName() string {
	return "GroupChatRenamed"
}

func (g *GroupChatRenamed) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatRenamed) GetName() *models2.GroupChatName {
	return &g.name
}

func (g *GroupChatRenamed) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatRenamed) GetExecutorId() *models2.UserAccountId {
	return &g.executorId
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
