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

func ValidateUserAccountId(id string) mo.Result[*models.UserAccountId] {
	return models.NewUserAccountIdFromString(id)
}
