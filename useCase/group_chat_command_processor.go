package useCase

import (
	"cqrs-es-example-go/domain"
	"cqrs-es-example-go/domain/events"
	"cqrs-es-example-go/domain/models"
	"cqrs-es-example-go/repository"
)

type GroupChatCommandProcessor struct {
	repository repository.GroupChatRepository
}

func NewGroupChatCommandProcessor(repository repository.GroupChatRepository) *GroupChatCommandProcessor {
	return &GroupChatCommandProcessor{
		repository,
	}
}

func (g *GroupChatCommandProcessor) CreateGroupChat(name *models.GroupChatName, administratorId *models.UserAccountId, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, event := domain.NewGroupChat(name, administratorId, executorId)
	err := g.repository.StoreEventWithSnapshot(event, groupChat)
	if err != nil {
		return nil, err
	}
	return event, nil
}
