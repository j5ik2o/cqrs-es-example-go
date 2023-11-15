package errors

type GroupChatRenameErr struct {
	GroupChatErr
}

func NewGroupChatRenameErr(message string) *GroupChatRenameErr {
	return &GroupChatRenameErr{GroupChatErr{message}}
}

func (e *GroupChatRenameErr) Error() string {
	return e.Message
}
