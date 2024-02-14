package api

type CreateGroupChatRequestBody struct {
	Name       string `json:"name" example:"group-chat-name-1"`
	ExecutorId string `json:"executor_id" example:"UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z"`
}

type CreateGroupChatResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id" example:"GroupChat-01H42K4ABWQ5V2XQEP3A48VE0Z"`
}

// ---

type DeleteGroupChatRequestBody struct {
	GroupChatId string `json:"group_chat_id"`
	ExecutorId  string `json:"executor_id"`
}

type DeleteGroupChatResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id"`
}

// ---

type RenameGroupChatRequestBody struct {
	GroupChatId string `json:"group_chat_id"`
	Name        string `json:"name"`
	ExecutorId  string `json:"executor_id"`
}

type RenameGroupChatResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id"`
}

// ---

type AddMemberRequestBody struct {
	GroupChatId   string `json:"group_chat_id"`
	UserAccountId string `json:"user_account_id"`
	Role          string `json:"role"`
	ExecutorId    string `json:"executor_id"`
}

type AddMemberResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id"`
}

// ---

type RemoveMemberRequestBody struct {
	GroupChatId   string `json:"group_chat_id"`
	UserAccountId string `json:"user_account_id"`
	ExecutorId    string `json:"executor_id"`
}

type RemoveMemberResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id"`
}

// ---

type PostMessageRequestBody struct {
	GroupChatId   string `json:"group_chat_id"`
	Message       string `json:"message"`
	UserAccountId string `json:"user_account_id"`
	ExecutorId    string `json:"executor_id"`
}

type PostMessageResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id"`
	MessageId   string `json:"message_id"`
}

// ---

type DeleteMessageRequestBody struct {
	GroupChatId   string `json:"group_chat_id"`
	MessageId     string `json:"message_id"`
	UserAccountId string `json:"user_account_id"`
	ExecutorId    string `json:"executor_id"`
}

type DeleteMessageResponseSuccessBody struct {
	GroupChatId string `json:"group_chat_id"`
}

// ---

type GroupChatResponseErrorBody struct {
	Message string `json:"message"`
}
