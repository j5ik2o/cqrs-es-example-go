package models

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

const GroupChatIdPrefix = "GroupChat"

// GroupChatId is a value object for group chat id.
type GroupChatId struct {
	value string
}

// NewGroupChatId is a constructor for GroupChatId
// It generates new GroupChatId
func NewGroupChatId() *GroupChatId {
	id := ulid.Make()
	return &GroupChatId{value: id.String()}
}

// NewGroupChatIdFromString is a constructor for GroupChatId
// It creates GroupChatId from string
func NewGroupChatIdFromString(value string) mo.Result[*GroupChatId] {
	if value == "" {
		return mo.Err[*GroupChatId](errors.New("GroupChatId is empty"))
	}
	if len(value) > len(GroupChatIdPrefix) && value[0:len(GroupChatIdPrefix)] == GroupChatIdPrefix {
		value = value[len(GroupChatIdPrefix)+1:]
	}
	return mo.Ok(&GroupChatId{value: value})
}

// ToJSON converts to JSON.
// ToJSON は JSON に変換します。
//
// However, this method is out of layer.
// ただし、このメソッドはレイヤーを逸脱しています。
func (g *GroupChatId) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"value": g.value,
	}
}

func (g *GroupChatId) GetValue() string {
	return g.value
}

func (g *GroupChatId) GetTypeName() string {
	return GroupChatIdPrefix
}

func (g *GroupChatId) AsString() string {
	return fmt.Sprintf("%s-%s", g.GetTypeName(), g.GetValue())
}

func (g *GroupChatId) String() string {
	return g.AsString()
}

// Equals compares other GroupChatId.
func (g *GroupChatId) Equals(other *GroupChatId) bool {
	return g.value == other.value
}

// ConvertGroupChatIdFromJSON converts from JSON.
func ConvertGroupChatIdFromJSON(value map[string]interface{}) mo.Result[*GroupChatId] {
	return NewGroupChatIdFromString(value["value"].(string))
}
