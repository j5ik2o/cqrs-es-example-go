package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatRenamed is a domain event for group chat renamed.
type GroupChatRenamed struct {
	id         string
	name       models.GroupChatName
	executorId models.UserAccountId
}

// NewGroupChatRenamed is a constructor for GroupChatRenamed with generating id.
func NewGroupChatRenamed(name models.GroupChatName, executorId models.UserAccountId) GroupChatRenamed {
	return GroupChatRenamed{ulid.Make().String(), name, executorId}
}

// NewGroupChatRenamedFrom is a constructor for GroupChatRenamed
func NewGroupChatRenamedFrom(id string, name models.GroupChatName, executorId models.UserAccountId) GroupChatRenamed {
	return GroupChatRenamed{id, name, executorId}
}

func (g *GroupChatRenamed) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"name":        g.name.ToJSON(),
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatRenamed) GetId() string {
	return g.id
}

func (g *GroupChatRenamed) GetTypeName() string {
	return EventTypeGroupChatRenamed
}

func (g *GroupChatRenamed) GetName() *models.GroupChatName {
	return &g.name
}

func (g *GroupChatRenamed) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatRenamed) String() string {
	return fmt.Sprintf("%s{ id: %s, name: %s }", g.GetTypeName(), g.id, g.name)
}
