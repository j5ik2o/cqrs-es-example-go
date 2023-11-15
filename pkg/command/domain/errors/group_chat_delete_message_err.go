package errors

type GroupChatDeleteMessageErr struct {
	GroupChatErr
}

func NewGroupChatDeleteMessageErr(message string) *GroupChatDeleteMessageErr {
	return &GroupChatDeleteMessageErr{GroupChatErr{message}}
}

func (e *GroupChatDeleteMessageErr) Error() string {
	return e.Message
}
