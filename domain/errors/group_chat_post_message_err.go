package errors

type GroupChatPostMessageErr struct {
	GroupChatErr
}

func NewGroupChatPostMessageErr(message string) *GroupChatPostMessageErr {
	return &GroupChatPostMessageErr{GroupChatErr{message}}
}

func (e *GroupChatPostMessageErr) Error() string {
	return e.Message
}
