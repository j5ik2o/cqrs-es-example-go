package useCase

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"github.com/samber/mo"
)

type GroupChatCommandProcessor struct {
	repository repository.GroupChatRepository
}

// NewGroupChatCommandProcessor is the constructor for GroupChatCommandProcessor.
func NewGroupChatCommandProcessor(repository repository.GroupChatRepository) GroupChatCommandProcessor {
	return GroupChatCommandProcessor{
		repository,
	}
}

// CreateGroupChat is the command handler for CreateGroupChat.
func (g *GroupChatCommandProcessor) CreateGroupChat(name models.GroupChatName, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, event := domain.NewGroupChat(name, executorId)
	errOpt := g.repository.StoreEventWithSnapshot(event, &groupChat)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(event)
}

// DeleteGroupChat is the command handler for DeleteGroupChat.
func (g *GroupChatCommandProcessor) DeleteGroupChat(groupChatId *models.GroupChatId, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	pair, err := groupChat.Delete(executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	errOpt := g.repository.StoreEventWithSnapshot(pair.V2, &pair.V1)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(pair.V2)
}

// RenameGroupChat is the command handler for RenameGroupChat.
func (g *GroupChatCommandProcessor) RenameGroupChat(groupChatId *models.GroupChatId, name models.GroupChatName, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	pair, err := groupChat.Rename(name, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	errOpt := g.repository.StoreEventWithSnapshot(pair.V2, &pair.V1)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(pair.V2)
}

// AddMember is the command handler for AddMember.
func (g *GroupChatCommandProcessor) AddMember(groupChatId *models.GroupChatId, userAccountId models.UserAccountId, role models.Role, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	memberId := models.NewMemberId()
	pair, err := groupChat.AddMember(memberId, userAccountId, role, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	errOpt := g.repository.StoreEventWithSnapshot(pair.V2, &pair.V1)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(pair.V2)
}

// RemoveMember is the command handler for RemoveMember.
func (g *GroupChatCommandProcessor) RemoveMember(groupChatId *models.GroupChatId, userAccountId models.UserAccountId, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	pair, err := groupChat.RemoveMemberByUserAccountId(userAccountId, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	errOpt := g.repository.StoreEventWithSnapshot(pair.V2, &pair.V1)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(pair.V2)
}

// PostMessage is the command handler for PostMessage.
func (g *GroupChatCommandProcessor) PostMessage(groupChatId *models.GroupChatId, message models.Message, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	pair, err := groupChat.PostMessage(message, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	errOpt := g.repository.StoreEventWithSnapshot(pair.V2, &pair.V1)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(pair.V2)
}

// DeleteMessage is the command handler for DeleteMessage.
func (g *GroupChatCommandProcessor) DeleteMessage(groupChatId *models.GroupChatId, messageId models.MessageId, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChat, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	pair, err := groupChat.DeleteMessage(messageId, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](err)
	}
	errOpt := g.repository.StoreEventWithSnapshot(pair.V2, &pair.V1)
	if err, b := errOpt.Get(); b {
		return mo.Err[events.GroupChatEvent](err)
	}
	return mo.Ok(pair.V2)
}
