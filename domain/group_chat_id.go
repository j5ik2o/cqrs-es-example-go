package domain

import (
	"fmt"
	"github.com/oklog/ulid/v2"
)

type GroupChatId struct {
	value string
}

func NewGroupChatId() *GroupChatId {
	id := ulid.Make()
	return &GroupChatId{value: id.String()}
}

func NewGroupChatIdFromString(value string) *GroupChatId {
	return &GroupChatId{value: value}
}

func (g *GroupChatId) GetValue() string {
	return g.value
}

func (g *GroupChatId) GetTypeName() string {
	return "group-chat"
}

func (g *GroupChatId) AsString() string {
	return fmt.Sprintf("%s-%s", g.GetTypeName(), g.GetValue())
}

func (g *GroupChatId) String() string {
	return g.AsString()
}
