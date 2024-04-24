package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
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

// GroupChatEvent is a domain event for group chat.
type GroupChatEvent interface {
	esa.Event
	GetExecutorId() *models.UserAccountId
	ToJSON() map[string]interface{}
}
