package domain

type GroupChatErr struct {
	message string
}

type GroupChatRenameErr struct {
	GroupChatErr
}

func NewGroupChatRenameErr(message string) *GroupChatRenameErr {
	return &GroupChatRenameErr{GroupChatErr{message}}
}

func (e *GroupChatErr) Error() string {
	return e.message
}

type GroupChatAddMemberErr struct {
	GroupChatErr
}

func NewGroupChatAddMemberErr(message string) *GroupChatAddMemberErr {
	return &GroupChatAddMemberErr{GroupChatErr{message}}
}

type GroupChatDeleteErr struct {
	GroupChatErr
}

func NewGroupChatDeleteErr(message string) *GroupChatDeleteErr {
	return &GroupChatDeleteErr{GroupChatErr{message}}
}

type GroupChatDeleteMessageErr struct {
	GroupChatErr
}

func NewGroupChatDeleteMessageErr(message string) *GroupChatDeleteMessageErr {
	return &GroupChatDeleteMessageErr{GroupChatErr{message}}
}

type GroupChatPostMessageErr struct {
	GroupChatErr
}

func NewGroupChatPostMessageErr(message string) *GroupChatPostMessageErr {
	return &GroupChatPostMessageErr{GroupChatErr{message}}
}

type GroupChatRemoveMemberErr struct {
	GroupChatErr
}

func NewGroupChatRemoveMemberErr(message string) *GroupChatRemoveMemberErr {
	return &GroupChatRemoveMemberErr{GroupChatErr{message}}
}
