package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatMessageDeleted is a domain event for group chat message deleted.
type GroupChatMessageDeleted struct {
	id         string
	messageId  models.MessageId
	executorId models.UserAccountId
}

// NewGroupChatMessageDeleted is a constructor for GroupChatMessageDeleted with generating id.
func NewGroupChatMessageDeleted(messageId models.MessageId, executorId models.UserAccountId) GroupChatMessageDeleted {
	return GroupChatMessageDeleted{ulid.Make().String(), messageId, executorId}
}

// NewGroupChatMessageDeletedFrom is a constructor for GroupChatMessageDeleted
func NewGroupChatMessageDeletedFrom(id string, messageId models.MessageId, executorId models.UserAccountId) GroupChatMessageDeleted {
	return GroupChatMessageDeleted{id, messageId, executorId}
}

func (g *GroupChatMessageDeleted) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"message_id":  g.messageId.ToJSON(),
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatMessageDeleted) GetId() string {
	return g.id
}

func (g *GroupChatMessageDeleted) GetTypeName() string {
	return EventTypeGroupChatMessageDeleted
}

func (g *GroupChatMessageDeleted) GetMessageId() *models.MessageId {
	return &g.messageId
}

func (g *GroupChatMessageDeleted) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMessageDeleted) String() string {
	return fmt.Sprintf("%s{ id: %s, messageId: %s }", g.GetTypeName(), g.id, g.messageId)
}
