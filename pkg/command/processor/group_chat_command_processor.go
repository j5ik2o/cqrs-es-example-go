package processor

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"github.com/samber/mo"
)

type CommandProcessError struct {
	message string
	Cause   error
}

func (c *CommandProcessError) Error() string {
	return c.message
}

type NotFoundError struct {
	CommandProcessError
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		CommandProcessError{
			message,
			nil,
		},
	}
}

type RepositoryError struct {
	CommandProcessError
}

func NewRepositoryError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		CommandProcessError{
			message,
			cause,
		},
	}
}

type DomainLogicError struct {
	CommandProcessError
}

func NewDomainLogicError(message string, cause error) *DomainLogicError {
	return &DomainLogicError{
		CommandProcessError{
			message,
			cause,
		},
	}
}

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
	if err, b := g.repository.Store(event, &groupChat).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}
	return mo.Ok(event)
}

// DeleteGroupChat is the command handler for DeleteGroupChat.
func (g *GroupChatCommandProcessor) DeleteGroupChat(groupChatId *models.GroupChatId, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChatOpt, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to find the group chat", err))
	}

	groupChat, b := groupChatOpt.Get()
	if !b {
		return mo.Err[events.GroupChatEvent](NewNotFoundError("The group chat is not found"))
	}

	pair, err := groupChat.Delete(executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewDomainLogicError("Failed to delete the group chat", err))
	}

	if err, b := g.repository.Store(pair.V2, &pair.V1).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}

	return mo.Ok(pair.V2)
}

// RenameGroupChat is the command handler for RenameGroupChat.
func (g *GroupChatCommandProcessor) RenameGroupChat(groupChatId *models.GroupChatId, name models.GroupChatName, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChatOpt, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to find the group chat", err))
	}

	groupChat, b := groupChatOpt.Get()
	if !b {
		return mo.Err[events.GroupChatEvent](NewNotFoundError("The group chat is not found"))
	}

	pair, err := groupChat.Rename(name, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewDomainLogicError("Failed to rename the group chat", err))
	}

	if err, b := g.repository.Store(pair.V2, &pair.V1).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}
	return mo.Ok(pair.V2)
}

// AddMember is the command handler for AddMember.
func (g *GroupChatCommandProcessor) AddMember(groupChatId *models.GroupChatId, userAccountId models.UserAccountId, role models.Role, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChatOpt, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to find the group chat", err))
	}

	groupChat, b := groupChatOpt.Get()
	if !b {
		return mo.Err[events.GroupChatEvent](NewNotFoundError("The group chat is not found"))
	}

	memberId := models.NewMemberId()
	pair, err := groupChat.AddMember(memberId, userAccountId, role, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewDomainLogicError("Failed to add the member to the group chat", err))
	}

	if err, b := g.repository.Store(pair.V2, &pair.V1).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}

	return mo.Ok(pair.V2)
}

// RemoveMember is the command handler for RemoveMember.
func (g *GroupChatCommandProcessor) RemoveMember(groupChatId *models.GroupChatId, userAccountId models.UserAccountId, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChatOpt, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to find the group chat", err))
	}

	groupChat, b := groupChatOpt.Get()
	if !b {
		return mo.Err[events.GroupChatEvent](NewNotFoundError("The group chat is not found"))
	}

	pair, err := groupChat.RemoveMemberByUserAccountId(userAccountId, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewDomainLogicError("Failed to remove the member from the group chat", err))
	}

	if err, b := g.repository.Store(pair.V2, &pair.V1).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}

	return mo.Ok(pair.V2)
}

// PostMessage is the command handler for PostMessage.
func (g *GroupChatCommandProcessor) PostMessage(groupChatId *models.GroupChatId, message models.Message, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChatOpt, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to find the group chat", err))
	}

	groupChat, b := groupChatOpt.Get()
	if !b {
		return mo.Err[events.GroupChatEvent](NewNotFoundError("The group chat is not found"))
	}

	pair, err := groupChat.PostMessage(message, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewDomainLogicError("Failed to post the message to the group chat", err))
	}

	if err, b := g.repository.Store(pair.V2, &pair.V1).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}

	return mo.Ok(pair.V2)
}

func (g *GroupChatCommandProcessor) EditMessage(groupChatId *models.GroupChatId, message models.Message, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	// TODO
	return mo.Err[events.GroupChatEvent](nil)
}

// DeleteMessage is the command handler for DeleteMessage.
func (g *GroupChatCommandProcessor) DeleteMessage(groupChatId *models.GroupChatId, messageId models.MessageId, executorId models.UserAccountId) mo.Result[events.GroupChatEvent] {
	groupChatOpt, err := g.repository.FindById(groupChatId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to find the group chat", err))
	}

	groupChat, b := groupChatOpt.Get()
	if !b {
		return mo.Err[events.GroupChatEvent](NewNotFoundError("The group chat is not found"))
	}

	pair, err := groupChat.DeleteMessage(messageId, executorId).Get()
	if err != nil {
		return mo.Err[events.GroupChatEvent](NewDomainLogicError("Failed to delete the message from the group chat", err))
	}

	if err, b := g.repository.Store(pair.V2, &pair.V1).Get(); b {
		return mo.Err[events.GroupChatEvent](NewRepositoryError("Failed to store the group chat", err))
	}

	return mo.Ok(pair.V2)
}
