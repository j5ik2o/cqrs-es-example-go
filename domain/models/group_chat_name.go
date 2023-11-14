package models

import "errors"

type GroupChatName struct {
	value string
}

func NewGroupChatName(value string) (*GroupChatName, error) {
	if value == "" {
		return nil, errors.New("GroupChatName is empty")
	}
	return &GroupChatName{value}, nil
}

func ConvertGroupChatNameFromJSON(value map[string]interface{}) (*GroupChatName, error) {
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
