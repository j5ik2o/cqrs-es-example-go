package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatCreated is a domain event for group chat created.
type GroupChatCreated struct {
	id         string
	name       models.GroupChatName
	members    models.Members
	executorId models.UserAccountId
}

// NewGroupChatCreated is a constructor for GroupChatCreated with generating id.
func NewGroupChatCreated(name models.GroupChatName, members models.Members, executorId models.UserAccountId) GroupChatCreated {
	return GroupChatCreated{ulid.Make().String(), name, members, executorId}
}

// NewGroupChatCreatedFrom is a constructor for GroupChatCreated
func NewGroupChatCreatedFrom(id string, name models.GroupChatName, members models.Members, executorId models.UserAccountId) GroupChatCreated {
	return GroupChatCreated{id, name, members, executorId}
}

func (g *GroupChatCreated) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"name":        g.name.ToJSON(),
		"members":     g.members.ToJSON(),
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatCreated) GetId() string {
	return g.id
}

func (g *GroupChatCreated) GetTypeName() string {
	return EventTypeGroupChatCreated
}

func (g *GroupChatCreated) GetName() *models.GroupChatName {
	return &g.name
}

func (g *GroupChatCreated) GetMembers() *models.Members {
	return &g.members
}

func (g *GroupChatCreated) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatCreated) String() string {
	return fmt.Sprintf("%s{ id: %s, name: %s }", g.GetTypeName(), g.id, g.name)
}
