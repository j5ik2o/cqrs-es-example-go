package errors

type GroupChatRemoveMemberErr struct {
	GroupChatErr
}

func NewGroupChatRemoveMemberErr(message string) *GroupChatRemoveMemberErr {
	return &GroupChatRemoveMemberErr{GroupChatErr{message}}
}

func (e *GroupChatRemoveMemberErr) Error() string {
	return e.Message
}
