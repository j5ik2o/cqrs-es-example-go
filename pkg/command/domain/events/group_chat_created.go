package events

import (
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/oklog/ulid/v2"
	"time"
)

type GroupChatCreated struct {
	id          string
	aggregateId *models2.GroupChatId
	name        *models2.GroupChatName
	members     *models2.Members
	seqNr       uint64
	executorId  *models2.UserAccountId
	occurredAt  uint64
}

func NewGroupChatCreated(aggregateId *models2.GroupChatId, name *models2.GroupChatName, members *models2.Members, seqNr uint64, executorId *models2.UserAccountId) *GroupChatCreated {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatCreated{id, aggregateId, name, members, seqNr, executorId, occurredAt}
}

func NewGroupChatCreatedFrom(id string, aggregateId *models2.GroupChatId, name *models2.GroupChatName, members *models2.Members, seqNr uint64, executorId *models2.UserAccountId, occurredAt uint64) *GroupChatCreated {
	return &GroupChatCreated{id, aggregateId, name, members, seqNr, executorId, occurredAt}
}

func (g *GroupChatCreated) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Id":          g.id,
		"AggregateId": g.aggregateId.ToJSON(),
		"Name":        g.name.ToJSON(),
		"Members":     g.members.ToJSON(),
		"SeqNr":       g.seqNr,
		"ExecutorId":  g.executorId.ToJSON(),
		"OccurredAt":  g.occurredAt,
	}
}

func (g *GroupChatCreated) GetId() string {
	return g.id
}

func (g *GroupChatCreated) GetTypeName() string {
	return "GroupChatCreated"
}

func (g *GroupChatCreated) GetAggregateId() esa.AggregateId {
	return g.aggregateId
}

func (g *GroupChatCreated) GetName() *models2.GroupChatName {
	return g.name
}

func (g *GroupChatCreated) GetMembers() *models2.Members {
	return g.members
}

func (g *GroupChatCreated) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatCreated) GetExecutorId() *models2.UserAccountId {
	return g.executorId
}

func (g *GroupChatCreated) IsCreated() bool {
	return true
}

func (g *GroupChatCreated) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatCreated) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s name: %s, seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.name, g.seqNr, g.occurredAt)
}
