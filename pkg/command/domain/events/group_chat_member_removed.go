package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatMemberRemoved is a domain event for group chat member removed.
type GroupChatMemberRemoved struct {
	id            string
	userAccountId models.UserAccountId
	executorId    models.UserAccountId
}

// NewGroupChatMemberRemoved is a constructor for GroupChatMemberRemoved with generating id.
func NewGroupChatMemberRemoved(userAccountId models.UserAccountId, executorId models.UserAccountId) GroupChatMemberRemoved {
	return GroupChatMemberRemoved{ulid.Make().String(), userAccountId, executorId}
}

// NewGroupChatMemberRemovedFrom is a constructor for GroupChatMemberRemoved
func NewGroupChatMemberRemovedFrom(id string, userAccountId models.UserAccountId, executorId models.UserAccountId) GroupChatMemberRemoved {
	return GroupChatMemberRemoved{id, userAccountId, executorId}
}

func (g *GroupChatMemberRemoved) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":            g.GetTypeName(),
		"id":              g.id,
		"user_account_id": g.userAccountId.ToJSON(),
		"executor_id":     g.executorId.ToJSON(),
	}
}

func (g *GroupChatMemberRemoved) GetId() string {
	return g.id
}

func (g *GroupChatMemberRemoved) GetTypeName() string {
	return EventTypeGroupChatMemberRemoved
}

func (g *GroupChatMemberRemoved) GetUserAccountId() *models.UserAccountId {
	return &g.userAccountId
}

func (g *GroupChatMemberRemoved) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMemberRemoved) String() string {
	return fmt.Sprintf("%s{ id: %s, userAccountId: %s }", g.GetTypeName(), g.id, g.userAccountId)
}
