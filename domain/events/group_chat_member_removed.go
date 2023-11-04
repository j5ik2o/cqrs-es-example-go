package events

import (
	"cqrs-es-example-go/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

type GroupChatMemberRemoved struct {
	id            string
	aggregateId   *models.GroupChatId
	seqNr         uint64
	userAccountId *models.UserAccountId
	executorId    *models.UserAccountId
	occurredAt    uint64
}

func NewGroupChatMemberRemoved(aggregateId *models.GroupChatId, seqNr uint64, userAccountId *models.UserAccountId, executorId *models.UserAccountId) *GroupChatMemberRemoved {
	id := ulid.Make()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatMemberRemoved{id: id.String(), aggregateId: aggregateId, seqNr: seqNr, userAccountId: userAccountId, executorId: executorId, occurredAt: occurredAt}
}

func (g *GroupChatMemberRemoved) GetId() string {
	return g.id
}

func (g *GroupChatMemberRemoved) GetTypeName() string {
	return "group-chat-member-removed"
}

func (g *GroupChatMemberRemoved) GetAggregateId() esa.AggregateId {
	return g.aggregateId
}

func (g *GroupChatMemberRemoved) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatMemberRemoved) GetUserAccountId() *models.UserAccountId {
	return g.userAccountId
}

func (g *GroupChatMemberRemoved) GetExecutorId() *models.UserAccountId {
	return g.executorId
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
