package models

import (
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

type GroupChatId struct {
	value string
}

func NewGroupChatId() *GroupChatId {
	id := ulid.Make()
	return &GroupChatId{value: id.String()}
}

func NewGroupChatIdFromString(value string) mo.Result[*GroupChatId] {
	// 先頭がGroupChat-であれば、それを削除する
	if len(value) > 9 && value[0:9] == "GroupChat" {
		value = value[10:]
	}
	return mo.Ok(&GroupChatId{value: value})
}

func (g *GroupChatId) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"value": g.value,
	}
}

func (g *GroupChatId) GetValue() string {
	return g.value
}

func (g *GroupChatId) GetTypeName() string {
	return "GroupChat"
}

func (g *GroupChatId) AsString() string {
	return fmt.Sprintf("%s-%s", g.GetTypeName(), g.GetValue())
}

func (g *GroupChatId) String() string {
	return g.AsString()
}

func ConvertGroupChatIdFromJSON(value map[string]interface{}) mo.Result[*GroupChatId] {
	return NewGroupChatIdFromString(value["value"].(string))
}
