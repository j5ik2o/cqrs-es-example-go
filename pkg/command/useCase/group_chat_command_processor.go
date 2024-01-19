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

func (g *GroupChatCommandProcessor) CreateGroupChat(name *models.GroupChatName, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, event := domain.NewGroupChat(name, executorId)
	err := g.repository.StoreEventWithSnapshot(event, groupChat)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (g *GroupChatCommandProcessor) DeleteGroupChat(groupChatId *models.GroupChatId, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return nil, err
	}
	pair, err := groupChat.Delete(executorId).Get()
	if err != nil {
		return nil, err
	}
	err = g.repository.StoreEventWithSnapshot(pair.V2, pair.V1)
	if err != nil {
		return nil, err
	}
	return pair.V2, nil
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

func (g *GroupChatCommandProcessor) AddMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId, role models.Role, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return nil, err
	}
	memberId := models.NewMemberId()
	pair, err := groupChat.AddMember(memberId, userAccountId, role, executorId).Get()
	if err != nil {
		return nil, err
	}
	err = g.repository.StoreEventWithSnapshot(pair.V2, pair.V1)
	if err != nil {
		return nil, err
	}
	return pair.V2, nil
}

func (g *GroupChatCommandProcessor) RemoveMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return nil, err
	}
	pair, err := groupChat.RemoveMemberByUserAccountId(userAccountId, executorId).Get()
	if err != nil {
		return nil, err
	}
	err = g.repository.StoreEventWithSnapshot(pair.V2, pair.V1)
	if err != nil {
		return nil, err
	}
	return pair.V2, nil
}

func (g *GroupChatCommandProcessor) PostMessage(groupChatId *models.GroupChatId, message *models.Message, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return nil, err
	}
	pair, err := groupChat.PostMessage(message, executorId).Get()
	if err != nil {
		return nil, err
	}
	err = g.repository.StoreEventWithSnapshot(pair.V2, pair.V1)
	if err != nil {
		return nil, err
	}
	return pair.V2, nil
}

func (g *GroupChatCommandProcessor) DeleteMessage(groupChatId *models.GroupChatId, messageId *models.MessageId, executorId *models.UserAccountId) (events.GroupChatEvent, error) {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return nil, err
	}
	pair, err := groupChat.DeleteMessage(messageId, executorId).Get()
	if err != nil {
		return nil, err
	}
	err = g.repository.StoreEventWithSnapshot(pair.V2, pair.V1)
	if err != nil {
		return nil, err
	}
	return pair.V2, nil
}
