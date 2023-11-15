package errors

type GroupChatAddMemberErr struct {
	GroupChatErr
}

func NewGroupChatAddMemberErr(message string) *GroupChatAddMemberErr {
	return &GroupChatAddMemberErr{GroupChatErr{message}}
}

func (e *GroupChatAddMemberErr) Error() string {
	return e.Message
}
