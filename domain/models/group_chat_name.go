package models

type GroupChatName struct {
	value string
}

func NewGroupChatName(value string) *GroupChatName {
	return &GroupChatName{value}
}

func ConvertGroupChatNameFromJSON(value map[string]interface{}) *GroupChatName {
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
