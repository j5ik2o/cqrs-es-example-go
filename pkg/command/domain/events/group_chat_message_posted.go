package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatMessagePosted is a domain event for group chat message posted.
type GroupChatMessagePosted struct {
	id         string
	message    models.Message
	executorId models.UserAccountId
}

// NewGroupChatMessagePosted is a constructor for GroupChatMessagePosted with generating id.
func NewGroupChatMessagePosted(message models.Message, executorId models.UserAccountId) GroupChatMessagePosted {
	return GroupChatMessagePosted{ulid.Make().String(), message, executorId}
}

// NewGroupChatMessagePostedFrom is a constructor for GroupChatMessagePosted
func NewGroupChatMessagePostedFrom(id string, message models.Message, executorId models.UserAccountId) GroupChatMessagePosted {
	return GroupChatMessagePosted{id, message, executorId}
}

func (g *GroupChatMessagePosted) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"message":     g.message.ToJSON(),
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatMessagePosted) GetId() string {
	return g.id
}

func (g *GroupChatMessagePosted) GetTypeName() string {
	return EventTypeGroupChatMessagePosted
}

func (g *GroupChatMessagePosted) GetMessage() *models.Message {
	return &g.message
}

func (g *GroupChatMessagePosted) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMessagePosted) String() string {
	return fmt.Sprintf("%s{ id: %s, message: %v }", g.GetTypeName(), g.id, g.message)
}
