package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatDeleted is a domain event for group chat deleted.
type GroupChatDeleted struct {
	id         string
	executorId models.UserAccountId
}

// NewGroupChatDeleted is a constructor for GroupChatDeleted with generating id.
func NewGroupChatDeleted(executorId models.UserAccountId) GroupChatDeleted {
	return GroupChatDeleted{ulid.Make().String(), executorId}
}

// NewGroupChatDeletedFrom is a constructor for GroupChatDeleted
func NewGroupChatDeletedFrom(id string, executorId models.UserAccountId) GroupChatDeleted {
	return GroupChatDeleted{id, executorId}
}

func (g *GroupChatDeleted) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatDeleted) GetId() string {
	return g.id
}

func (g *GroupChatDeleted) GetTypeName() string {
	return EventTypeGroupChatDeleted
}

func (g *GroupChatDeleted) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatDeleted) String() string {
	return fmt.Sprintf("%s{ id: %s }", g.GetTypeName(), g.id)
}
