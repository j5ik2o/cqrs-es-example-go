package ctrl

import (
	"cqrs-es-example-go/domain/models"
	"github.com/samber/mo"
)

func ValidateGroupChatName(name string) mo.Result[*models.GroupChatName] {
	return models.NewGroupChatName(name)
}

func ValidateUserAccountId(id string) mo.Result[*models.UserAccountId] {
	return models.NewUserAccountIdFromString(id)
}
