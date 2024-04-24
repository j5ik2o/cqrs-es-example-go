package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
	"github.com/oklog/ulid/v2"
	"time"
)

// GroupChatMessagePosted is a domain event for group chat message posted.
type GroupChatMessagePosted struct {
	id          string
	aggregateId models.GroupChatId
	message     models.Message
	seqNr       uint64
	executorId  models.UserAccountId
	occurredAt  uint64
}

// NewGroupChatMessagePosted is a constructor for GroupChatMessagePosted with generating id.
func NewGroupChatMessagePosted(aggregateId models.GroupChatId, message models.Message, seqNr uint64, executorId models.UserAccountId) GroupChatMessagePosted {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return GroupChatMessagePosted{id, aggregateId, message, seqNr, executorId, occurredAt}
}

// NewGroupChatMessagePostedFrom is a constructor for GroupChatMessagePosted
func NewGroupChatMessagePostedFrom(id string, aggregateId models.GroupChatId, message models.Message, seqNr uint64, executorId models.UserAccountId, occurredAt uint64) GroupChatMessagePosted {
	return GroupChatMessagePosted{id, aggregateId, message, seqNr, executorId, occurredAt}
}

func (g *GroupChatMessagePosted) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type_name":    g.GetTypeName(),
		"id":           g.id,
		"aggregate_id": g.aggregateId.ToJSON(),
		"message":      g.message.ToJSON(),
		"executor_id":  g.executorId.ToJSON(),
		"seq_nr":       g.seqNr,
		"occurred_at":  g.occurredAt,
	}
}

func (g *GroupChatMessagePosted) GetId() string {
	return g.id
}

func (g *GroupChatMessagePosted) GetTypeName() string {
	return "GroupChatMessagePosted"
}

func (g *GroupChatMessagePosted) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatMessagePosted) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatMessagePosted) GetMessage() *models.Message {
	return &g.message
}

func (g *GroupChatMessagePosted) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMessagePosted) IsCreated() bool {
	return false
}

func (g *GroupChatMessagePosted) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatMessagePosted) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.seqNr, g.occurredAt)
}
