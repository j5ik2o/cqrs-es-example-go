package events

import (
	"cqrs-es-example-go/domain/models"
	"fmt"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"time"
)

type GroupChatPosted struct {
	id          string
	aggregateId *models.GroupChatId
	seqNr       uint64
	message     *models.Message
	executorId  *models.UserAccountId
	occurredAt  uint64
}

func NewGroupChatPosted(id string, aggregateId *models.GroupChatId, seqNr uint64, message *models.Message, executorId *models.UserAccountId) *GroupChatPosted {
	now := time.Now()
	occurredAt := uint64(now.UnixNano() / 1e6)
	return &GroupChatPosted{id, aggregateId, seqNr, message, executorId, occurredAt}
}

func (g *GroupChatPosted) GetId() string {
	return g.id
}

func (g *GroupChatPosted) GetTypeName() string {
	return "group-chat-posted"
}

func (g *GroupChatPosted) GetAggregateId() esa.AggregateId {
	return g.aggregateId
}

func (g *GroupChatPosted) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChatPosted) GetMessage() *models.Message {
	return g.message
}

func (g *GroupChatPosted) GetExecutorId() *models.UserAccountId {
	return g.executorId
}

func (g *GroupChatPosted) IsCreated() bool {
	return false
}

func (g *GroupChatPosted) GetOccurredAt() uint64 {
	return g.occurredAt
}

func (g *GroupChatPosted) String() string {
	return fmt.Sprintf("%s{ id: %s, aggregateId: %s seqNr: %d, occurredAt: %d}",
		g.GetTypeName(), g.id, g.aggregateId, g.seqNr, g.occurredAt)
}
