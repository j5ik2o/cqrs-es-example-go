package validator

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"github.com/samber/mo"
)

func ValidateGroupChatId(id string) mo.Result[*models.GroupChatId] {
	return models.NewGroupChatIdFromString(id)
}

func ValidateGroupChatName(name string) mo.Result[*models.GroupChatName] {
	return models.NewGroupChatName(name)
}

func ValidateUserAccountId(id string) mo.Result[models.UserAccountId] {
	return models.NewUserAccountIdFromString(id)
}

func ValidateMessage(id *models.MessageId, message string, senderId models.UserAccountId) mo.Result[*models.Message] {
	return models.NewMessage(id, message, senderId)
}

func ValidateMessageId(id string) mo.Result[*models.MessageId] {
	return models.NewMessageIdFromString(id)
}
