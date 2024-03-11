package errors

type GroupChatError struct {
	message string
}

func (e *GroupChatError) Error() string {
	return e.message
}

// ---

type AlreadyDeletedError struct {
	GroupChatError
}

func NewAlreadyDeletedError(message string) *AlreadyDeletedError {
	return &AlreadyDeletedError{GroupChatError{message}}
}

// ---

type AlreadyExistsMessageError struct {
	GroupChatError
}

func NewAlreadyExistsMessageError(message string) *AlreadyExistsMessageError {
	return &AlreadyExistsMessageError{GroupChatError{message}}
}

// ---

type NotMemberError struct {
	GroupChatError
}

func NewNotMemberError(message string) *NotMemberError {
	return &NotMemberError{GroupChatError{message}}
}

// ---

type NotAdministratorError struct {
	GroupChatError
}

func NewNotAdministratorError(message string) *NotAdministratorError {
	return &NotAdministratorError{GroupChatError{message}}
}

// ---

type MessageNotFoundError struct {
	GroupChatError
}

func NewMessageNotFoundError(message string) *MessageNotFoundError {
	return &MessageNotFoundError{GroupChatError{message}}
}

// ---

type NotSenderError struct {
	GroupChatError
}

func NewNotSenderError(message string) *NotSenderError {
	return &NotSenderError{GroupChatError{message}}
}

// ---

type AlreadyExistsNameError struct {
	GroupChatError
}

func NewAlreadyExistsNameError(message string) *AlreadyExistsNameError {
	return &AlreadyExistsNameError{GroupChatError{message}}
}

// ---

type MismatchedUserAccountError struct {
	GroupChatError
}

func NewMismatchedUserAccountError(message string) *MismatchedUserAccountError {
	return &MismatchedUserAccountError{GroupChatError{message}}
}
