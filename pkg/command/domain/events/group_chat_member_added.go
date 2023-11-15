package events

import (
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

type GroupChatMemberAdded struct {
	id          string
	aggregateId *models2.GroupChatId
	member      *models2.Member
	seqNr       uint64
	executorId  *models2.UserAccountId
	occurredAt  uint64
}

func NewGroupChatMemberAdded(aggregateId *models2.GroupChatId, member *models2.Member, seqNr uint64, executorId *models2.UserAccountId) *GroupChatMemberAdded {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatMemberAdded{id, aggregateId, member, seqNr, executorId, occurredAt}
}

func NewGroupChatMemberAddedFrom(id string, aggregateId *models2.GroupChatId, member *models2.Member, seqNr uint64, executorId *models2.UserAccountId, occurredAt uint64) *GroupChatMemberAdded {
	return &GroupChatMemberAdded{id, aggregateId, member, seqNr, executorId, occurredAt}
}

func (g *GroupChatMemberAdded) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Id":          g.id,
		"AggregateId": g.aggregateId.ToJSON(),
		"Member":      g.member.ToJSON(),
		"SeqNr":       g.seqNr,
		"ExecutorId":  g.executorId.ToJSON(),
		"OccurredAt":  g.occurredAt,
	}
}

func (g *GroupChatMemberAdded) GetId() string {
	return g.id
}

func (g *GroupChatMemberAdded) GetTypeName() string {
	return "GroupChatMemberAdded"
}

func (g *GroupChatMemberAdded) GetAggregateId() esa.AggregateId {
	return g.aggregateId
}

func (g *GroupChatMemberAdded) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatMemberAdded) GetMember() *models2.Member {
	return g.member
}

func (g *GroupChatMemberAdded) GetExecutorId() *models2.UserAccountId {
	return g.executorId
}

func (g *GroupChatMemberAdded) IsCreated() bool {
	return false
}

func (g *GroupChatMemberAdded) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatMemberAdded) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s, seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.seqNr, g.occurredAt)
}
