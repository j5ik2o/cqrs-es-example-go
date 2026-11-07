package events

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	"github.com/oklog/ulid/v2"
)

// GroupChatMemberAdded is a domain event for group chat member added.
type GroupChatMemberAdded struct {
	id         string
	member     models.Member
	executorId models.UserAccountId
}

// NewGroupChatMemberAdded is a constructor for GroupChatMemberAdded with generating id.
func NewGroupChatMemberAdded(member models.Member, executorId models.UserAccountId) GroupChatMemberAdded {
	return GroupChatMemberAdded{ulid.Make().String(), member, executorId}
}

// NewGroupChatMemberAddedFrom is a constructor for GroupChatMemberAdded
func NewGroupChatMemberAddedFrom(id string, member models.Member, executorId models.UserAccountId) GroupChatMemberAdded {
	return GroupChatMemberAdded{id, member, executorId}
}

func (g *GroupChatMemberAdded) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        g.GetTypeName(),
		"id":          g.id,
		"member":      g.member.ToJSON(),
		"executor_id": g.executorId.ToJSON(),
	}
}

func (g *GroupChatMemberAdded) GetId() string {
	return g.id
}

func (g *GroupChatMemberAdded) GetTypeName() string {
	return EventTypeGroupChatMemberAdded
}

func (g *GroupChatMemberAdded) GetMember() *models.Member {
	return &g.member
}

func (g *GroupChatMemberAdded) GetExecutorId() *models.UserAccountId {
	return &g.executorId
}

func (g *GroupChatMemberAdded) String() string {
	return fmt.Sprintf("%s{ id: %s, member: %v }", g.GetTypeName(), g.id, g.member)
}
