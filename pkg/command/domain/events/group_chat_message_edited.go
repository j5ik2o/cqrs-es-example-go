package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatMessageEdited is a domain event for group chat message edited.
type GroupChatMessageEdited struct {
	id         string
	message    models.Message
	executorId models.UserAccountId
}

// NewGroupChatMessageEdited is a constructor for GroupChatMessageEdited with generating id.
func NewGroupChatMessageEdited(message models.Message, executorId models.UserAccountId) GroupChatMessageEdited {
	return GroupChatMessageEdited{ulid.Make().String(), message, executorId}
}

// NewGroupChatMessageEditedFrom is a constructor for GroupChatMessageEdited
func NewGroupChatMessageEditedFrom(id string, message models.Message, executorId models.UserAccountId) GroupChatMessageEdited {
	return GroupChatMessageEdited{id, message, executorId}
}

func (g *GroupChatMessageEdited) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"message":     g.message.ToJSON(),
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatMessageEdited) GetId() string {
	return g.id
}

func (g *GroupChatMessageEdited) GetTypeName() string {
	return EventTypeGroupChatMessageEdited
}

func (g *GroupChatMessageEdited) GetMessage() *models.Message {
	return &g.message
}

func (g *GroupChatMessageEdited) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMessageEdited) String() string {
	return fmt.Sprintf("%s{ id: %s, message: %v }", g.GetTypeName(), g.id, g.message)
}
