package models

type GroupChatName struct {
	value string
}

func NewGroupChatName(value string) *GroupChatName {
	return &GroupChatName{value}
}

func (g GroupChatName) Sting() string {
	return g.value
}
