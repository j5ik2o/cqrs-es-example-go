package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

// GroupChatRenamed is a domain event for group chat renamed.
type GroupChatRenamed struct {
	id          string
	aggregateId models.GroupChatId
	name        models.GroupChatName
	seqNr       uint64
	executorId  models.UserAccountId
	occurredAt  uint64
}

// NewGroupChatRenamed is a constructor for GroupChatRenamed with generating id.
func NewGroupChatRenamed(aggregateId models.GroupChatId, name models.GroupChatName, seqNr uint64, executorId models.UserAccountId) GroupChatRenamed {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return GroupChatRenamed{id, aggregateId, name, seqNr, executorId, occurredAt}
}

// NewGroupChatRenamedFrom is a constructor for GroupChatRenamed
func NewGroupChatRenamedFrom(id string, aggregateId models.GroupChatId, name models.GroupChatName, seqNr uint64, executorId models.UserAccountId, occurredAt uint64) GroupChatRenamed {
	return GroupChatRenamed{id, aggregateId, name, seqNr, executorId, occurredAt}
}

func (g *GroupChatRenamed) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type_name":    g.GetTypeName(),
		"id":           g.id,
		"aggregate_id": g.aggregateId.ToJSON(),
		"name":         g.name.ToJSON(),
		"executor_id":  g.executorId.ToJSON(),
		"seq_nr":       g.seqNr,
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

func (g *GroupChatRenamed) GetName() *models.GroupChatName {
	return &g.name
}

func (g *GroupChatRenamed) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatRenamed) GetExecutorId() *models.UserAccountId {
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
