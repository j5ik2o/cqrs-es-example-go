package models

import (
	"errors"
	"github.com/samber/mo"
)

// GroupChatName is a value object that represents a group chat name.
type GroupChatName struct {
	value string
}

// NewGroupChatName is the constructor for GroupChatName.
func NewGroupChatName(value string) mo.Result[GroupChatName] {
	if value == "" {
		return mo.Err[GroupChatName](errors.New("GroupChatName is empty"))
	}
	return mo.Ok(GroupChatName{value})
}

// ConvertGroupChatNameFromJSON is a constructor for GroupChatName.
func ConvertGroupChatNameFromJSON(value map[string]interface{}) mo.Result[GroupChatName] {
	return NewGroupChatName(value["value"].(string))
}

// ToJSON converts to JSON.
//
// However, this method is out of layer.
func (g *GroupChatName) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"value": g.value,
	}
}

func (g *GroupChatName) String() string {
	return g.value
}

func (g *GroupChatName) Equals(other *GroupChatName) bool {
	return g.value == other.value
}
