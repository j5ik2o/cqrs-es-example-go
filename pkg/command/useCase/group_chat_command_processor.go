package useCase

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
)

type GroupChatCommandProcessor struct {
	repository repository.GroupChatRepository
}

func NewGroupChatCommandProcessor(repository repository.GroupChatRepository) *GroupChatCommandProcessor {
	return &GroupChatCommandProcessor{
		repository,
	}
}

func (g *GroupChatCommandProcessor) CreateGroupChat(name *models2.GroupChatName, administratorId *models2.UserAccountId, executorId *models2.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, event := domain.NewGroupChat(name, administratorId, executorId)
	err := g.repository.StoreEventWithSnapshot(event, groupChat)
	if err != nil {
		return nil, err
	}
	return event, nil
}
