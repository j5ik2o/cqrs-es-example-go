package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

// GroupChatDeleted is a domain event for group chat deleted.
type GroupChatDeleted struct {
	id          string
	aggregateId models.GroupChatId
	seqNr       uint64
	executorId  models.UserAccountId
	occurredAt  uint64
}

// NewGroupChatDeleted is a constructor for GroupChatDeleted with generating id.
func NewGroupChatDeleted(aggregateId models.GroupChatId, seqNr uint64, executorId models.UserAccountId) GroupChatDeleted {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return GroupChatDeleted{id, aggregateId, seqNr, executorId, occurredAt}
}

// NewGroupChatDeletedFrom is a constructor for GroupChatDeleted
func NewGroupChatDeletedFrom(id string, aggregateId models.GroupChatId, seqNr uint64, executorId models.UserAccountId, occurredAt uint64) GroupChatDeleted {
	return GroupChatDeleted{id, aggregateId, seqNr, executorId, occurredAt}
}

func (g *GroupChatDeleted) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type_name":    g.GetTypeName(),
		"id":           g.id,
		"aggregate_id": g.aggregateId.ToJSON(),
		"executor_id":  g.executorId.ToJSON(),
		"seq_nr":       g.seqNr,
		"occurred_at":  g.occurredAt,
	}
}

func (g *GroupChatDeleted) GetId() string {
	return g.id
}

func (g *GroupChatDeleted) GetTypeName() string {
	return "GroupChatDeleted"
}

func (g *GroupChatDeleted) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatDeleted) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatDeleted) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatDeleted) IsCreated() bool {
	return false
}

func (g *GroupChatDeleted) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatDeleted) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.seqNr, g.occurredAt)
}
