package events

import (
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

type GroupChatMemberRemoved struct {
	id            string
	aggregateId   models2.GroupChatId
	userAccountId models2.UserAccountId
	seqNr         uint64
	executorId    models2.UserAccountId
	occurredAt    uint64
}

func NewGroupChatMemberRemoved(aggregateId models2.GroupChatId, userAccountId models2.UserAccountId, seqNr uint64, executorId models2.UserAccountId) *GroupChatMemberRemoved {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatMemberRemoved{id, aggregateId, userAccountId, seqNr, executorId, occurredAt}
}

func NewGroupChatMemberRemovedFrom(id string, aggregateId models2.GroupChatId, userAccountId models2.UserAccountId, seqNr uint64, executorId models2.UserAccountId, occurredAt uint64) *GroupChatMemberRemoved {
	return &GroupChatMemberRemoved{id, aggregateId, userAccountId, seqNr, executorId, occurredAt}
}

func (g *GroupChatMemberRemoved) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":             g.id,
		"aggregate_id":   g.aggregateId.ToJSON(),
		"userAccount_id": g.userAccountId.ToJSON(),
		"seq_nr":         g.seqNr,
		"executor_id":    g.executorId.ToJSON(),
		"occurred_at":    g.occurredAt,
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

func (g *GroupChatMemberRemoved) GetUserAccountId() *models2.UserAccountId {
	return &g.userAccountId
}

func (g *GroupChatMemberRemoved) GetExecutorId() *models2.UserAccountId {
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
