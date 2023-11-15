package useCase

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
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

func (g *GroupChatCommandProcessor) CreateGroupChat(name *models.GroupChatName, administratorId *models.UserAccountId, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, event := domain.NewGroupChat(name, administratorId, executorId)
	err := g.repository.StoreEventWithSnapshot(event, groupChat)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (g *GroupChatCommandProcessor) RenameGroupChat(groupChatId *models.GroupChatId, name *models.GroupChatName, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return nil, err
	}
	pair, err := groupChat.Rename(name, executorId).Get()
	if err != nil {
		return nil, err
	}
	err = g.repository.StoreEventWithSnapshot(pair.V2, pair.V1)
	if err != nil {
		return nil, err
	}
	return pair.V2, nil
}
