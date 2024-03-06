package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

// GroupChatMessageEdited is a domain event for group chat message posted.
type GroupChatMessageEdited struct {
	id          string
	aggregateId models.GroupChatId
	message     models.Message
	seqNr       uint64
	executorId  models.UserAccountId
	occurredAt  uint64
}

// NewGroupChatMessageEdited is a constructor for GroupChatMessageEdited with generating id.
func NewGroupChatMessageEdited(aggregateId models.GroupChatId, message models.Message, seqNr uint64, executorId models.UserAccountId) GroupChatMessageEdited {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return GroupChatMessageEdited{id, aggregateId, message, seqNr, executorId, occurredAt}
}

// NewGroupChatMessageEditedFrom is a constructor for GroupChatMessageEdited
func NewGroupChatMessageEditedFrom(id string, aggregateId models.GroupChatId, message models.Message, seqNr uint64, executorId models.UserAccountId, occurredAt uint64) GroupChatMessageEdited {
	return GroupChatMessageEdited{id, aggregateId, message, seqNr, executorId, occurredAt}
}

func (g *GroupChatMessageEdited) ToJSON() map[string]interface{} {
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

func (g *GroupChatMessageEdited) GetId() string {
	return g.id
}

func (g *GroupChatMessageEdited) GetTypeName() string {
	return "GroupChatMessageEdited"
}

func (g *GroupChatMessageEdited) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatMessageEdited) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatMessageEdited) GetMessage() *models.Message {
	return &g.message
}

func (g *GroupChatMessageEdited) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMessageEdited) IsCreated() bool {
	return false
}

func (g *GroupChatMessageEdited) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatMessageEdited) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s, message: %s , executorId: %s, seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.message, g.executorId, g.seqNr, g.occurredAt)
}
