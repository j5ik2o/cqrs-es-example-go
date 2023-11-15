package errors

type GroupChatErr struct {
	Message string
}

func (e *GroupChatErr) Error() string {
	return e.Message
}
