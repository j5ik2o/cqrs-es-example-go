package rmu

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"time"
)

type GroupChatDao interface {
	InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error
	DeleteGroupChat(aggregateId *models.GroupChatId, updatedAt time.Time) error
	UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, updatedAt time.Time) error

	InsertMember(aggregateId *models.GroupChatId, member *models.Member, createdAt time.Time) error
	DeleteMember(aggregateId *models.GroupChatId, userAccountId *models.UserAccountId) error

	InsertMessage(messageId *models.MessageId, aggregateId *models.GroupChatId, userAccountId *models.UserAccountId, text string, createdAt time.Time) error
	UpdateMessage(messageId *models.MessageId, text string, updatedAt time.Time) error
	DeleteMessage(messageId *models.MessageId, updatedAt time.Time) error
}
