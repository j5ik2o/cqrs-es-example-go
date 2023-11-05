package errors

type GroupChatDeleteErr struct {
	GroupChatErr
}

func NewGroupChatDeleteErr(message string) *GroupChatDeleteErr {
	return &GroupChatDeleteErr{GroupChatErr{message}}
}

func (e *GroupChatDeleteErr) Error() string {
	return e.Message
}
