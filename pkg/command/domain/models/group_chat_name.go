package models

import (
	"errors"
	"github.com/samber/mo"
)

type GroupChatName struct {
	value string
}

func NewGroupChatName(value string) mo.Result[*GroupChatName] {
	if value == "" {
		return mo.Err[*GroupChatName](errors.New("GroupChatName is empty"))
	}
	return mo.Ok(&GroupChatName{value})
}

func ConvertGroupChatNameFromJSON(value map[string]interface{}) mo.Result[*GroupChatName] {
	return NewGroupChatName(value["Value"].(string))
}

func (g *GroupChatName) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Value": g.value,
	}
}

func (g *GroupChatName) String() string {
	return g.value
}
