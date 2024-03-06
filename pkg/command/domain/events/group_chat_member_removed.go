package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

// GroupChatMemberRemoved is a domain event for group chat member removed.
type GroupChatMemberRemoved struct {
	id            string
	aggregateId   models.GroupChatId
	userAccountId models.UserAccountId
	seqNr         uint64
	executorId    models.UserAccountId
	occurredAt    uint64
}

// NewGroupChatMemberRemoved is a constructor for GroupChatMemberRemoved with generating id.
func NewGroupChatMemberRemoved(aggregateId models.GroupChatId, userAccountId models.UserAccountId, seqNr uint64, executorId models.UserAccountId) GroupChatMemberRemoved {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return GroupChatMemberRemoved{id, aggregateId, userAccountId, seqNr, executorId, occurredAt}
}

// NewGroupChatMemberRemovedFrom is a constructor for GroupChatMemberRemoved
func NewGroupChatMemberRemovedFrom(id string, aggregateId models.GroupChatId, userAccountId models.UserAccountId, seqNr uint64, executorId models.UserAccountId, occurredAt uint64) GroupChatMemberRemoved {
	return GroupChatMemberRemoved{id, aggregateId, userAccountId, seqNr, executorId, occurredAt}
}

func (g *GroupChatMemberRemoved) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type_name":       g.GetTypeName(),
		"id":              g.id,
		"aggregate_id":    g.aggregateId.ToJSON(),
		"user_account_id": g.userAccountId.ToJSON(),
		"executor_id":     g.executorId.ToJSON(),
		"seq_nr":          g.seqNr,
		"occurred_at":     g.occurredAt,
	}
}

func (g *GroupChatMemberRemoved) GetId() string {
	return g.id
}

func (g *GroupChatMemberRemoved) GetTypeName() string {
	return "GroupChatMemberRemoved"
}

func (g *GroupChatMemberRemoved) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatMemberRemoved) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatMemberRemoved) GetUserAccountId() *models.UserAccountId {
	return &g.userAccountId
}

func (g *GroupChatMemberRemoved) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMemberRemoved) IsCreated() bool {
	return false
}

func (g *GroupChatMemberRemoved) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatMemberRemoved) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s, seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.seqNr, g.occurredAt)
}
