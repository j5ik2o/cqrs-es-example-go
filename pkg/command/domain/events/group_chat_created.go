package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
	"github.com/oklog/ulid/v2"
	"time"
)

// GroupChatCreated is a domain event for group chat created.
type GroupChatCreated struct {
	id          string
	aggregateId models.GroupChatId
	name        models.GroupChatName
	members     models.Members
	seqNr       uint64
	executorId  models.UserAccountId
	occurredAt  uint64
}

// NewGroupChatCreated is a constructor for GroupChatCreated with generating id.
func NewGroupChatCreated(aggregateId models.GroupChatId, name models.GroupChatName, members models.Members, seqNr uint64, executorId models.UserAccountId) GroupChatCreated {
	id := ulid.Make().String()
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return GroupChatCreated{id, aggregateId, name, members, seqNr, executorId, occurredAt}
}

// NewGroupChatCreatedFrom is a constructor for GroupChatCreated
func NewGroupChatCreatedFrom(id string, aggregateId models.GroupChatId, name models.GroupChatName, members models.Members, seqNr uint64, executorId models.UserAccountId, occurredAt uint64) GroupChatCreated {
	return GroupChatCreated{id, aggregateId, name, members, seqNr, executorId, occurredAt}
}

func (g *GroupChatCreated) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type_name":    g.GetTypeName(),
		"id":           g.id,
		"aggregate_id": g.aggregateId.ToJSON(),
		"name":         g.name.ToJSON(),
		"members":      g.members.ToJSON(),
		"executor_id":  g.executorId.ToJSON(),
		"seq_nr":       g.seqNr,
		"occurred_at":  g.occurredAt,
	}
}

func (g *GroupChatCreated) GetId() string {
	return g.id
}

func (g *GroupChatCreated) GetTypeName() string {
	return "GroupChatCreated"
}

func (g *GroupChatCreated) GetAggregateId() esa.AggregateId {
	return &g.aggregateId
}

func (g *GroupChatCreated) GetName() *models.GroupChatName {
	return &g.name
}

func (g *GroupChatCreated) GetMembers() *models.Members {
	return &g.members
}

func (g *GroupChatCreated) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatCreated) GetExecutorId() *models.UserAccountId {
	return &g.executorId
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
