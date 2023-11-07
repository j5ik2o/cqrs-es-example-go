package events

import (
	"cqrs-es-example-go/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

type GroupChatEvent interface {
	esa.Event
	GetExecutorId() *models.UserAccountId
	ToJSON() map[string]interface{}
}
