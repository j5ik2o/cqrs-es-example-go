package ctrl

import "cqrs-es-example-go/domain/models"

func ValidateGroupChatName(name string) (*models.GroupChatName, error) {
	return models.NewGroupChatName(name)
}

func ValidateUserAccountId(id string) (*models.UserAccountId, error) {
	return models.NewUserAccountIdFromString(id)
}
