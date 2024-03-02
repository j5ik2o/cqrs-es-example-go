package domain

type GroupChatError struct {
	message string
}

type GroupChatRenameError struct {
	GroupChatError
}

func NewGroupChatRenameError(message string) *GroupChatRenameError {
	return &GroupChatRenameError{GroupChatError{message}}
}

func (e *GroupChatError) Error() string {
	return e.message
}

type GroupChatAddMemberError struct {
	GroupChatError
}

func NewGroupChatAddMemberError(message string) *GroupChatAddMemberError {
	return &GroupChatAddMemberError{GroupChatError{message}}
}

type GroupChatDeleteError struct {
	GroupChatError
}

func NewGroupChatDeleteError(message string) *GroupChatDeleteError {
	return &GroupChatDeleteError{GroupChatError{message}}
}

type GroupChatDeleteMessageError struct {
	GroupChatError
}

func NewGroupChatDeleteMessageError(message string) *GroupChatDeleteMessageError {
	return &GroupChatDeleteMessageError{GroupChatError{message}}
}

type GroupChatPostMessageError struct {
	GroupChatError
}

func NewGroupChatPostMessageError(message string) *GroupChatPostMessageError {
	return &GroupChatPostMessageError{GroupChatError{message}}
}

type GroupChatRemoveMemberError struct {
	GroupChatError
}

func NewGroupChatRemoveMemberError(message string) *GroupChatRemoveMemberError {
	return &GroupChatRemoveMemberError{GroupChatError{message}}
}
