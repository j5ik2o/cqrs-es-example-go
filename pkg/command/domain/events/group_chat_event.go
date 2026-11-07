package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go/v5"
)

const (
	EventTypeGroupChatCreated        = "GroupChatCreated"
	EventTypeGroupChatDeleted        = "GroupChatDeleted"
	EventTypeGroupChatRenamed        = "GroupChatRenamed"
	EventTypeGroupChatMemberAdded    = "GroupChatMemberAdded"
	EventTypeGroupChatMemberRemoved  = "GroupChatMemberRemoved"
	EventTypeGroupChatMessagePosted  = "GroupChatMessagePosted"
	EventTypeGroupChatMessageDeleted = "GroupChatMessageDeleted"
)

// AggregateId is the aggregate identifier carried by a domain event.
// It also satisfies esa.AggregateID so it can be placed in an event envelope.
type AggregateId interface {
	esa.AggregateID
	GetTypeName() string
	GetValue() string
	AsString() string
}

// GroupChatEvent is a domain event for group chat.
type GroupChatEvent interface {
	GetId() string
	GetTypeName() string
	GetAggregateId() AggregateId
	GetSeqNr() uint64
	GetOccurredAt() uint64
	IsCreated() bool
	GetExecutorId() *models.UserAccountId
	ToJSON() map[string]interface{}
}
