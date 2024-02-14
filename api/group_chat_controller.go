package api

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/validator"
	"cqrs-es-example-go/pkg/command/useCase"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

// GroupChatController is a controller for group chat commands.
type GroupChatController struct {
	groupChatCommandProcessor useCase.GroupChatCommandProcessor
}

// NewGroupChatController creates a new GroupChatController.
func NewGroupChatController(groupChatCommandProcessor useCase.GroupChatCommandProcessor) GroupChatController {
	return GroupChatController{
		groupChatCommandProcessor,
	}
}

// CreateGroupChat is a handler for creating a group chat.
func (g *GroupChatController) CreateGroupChat(c *gin.Context) {
	var jsonRequestBody CreateGroupChatRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatName, err := validator.ValidateGroupChatName(jsonRequestBody.Name).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.CreateGroupChat(groupChatName, executorId)
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := CreateGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// DeleteGroupChat is a handler for deleting a group chat.
func (g *GroupChatController) DeleteGroupChat(c *gin.Context) {
	var jsonRequestBody DeleteGroupChatRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatId, err := validator.ValidateGroupChatId(jsonRequestBody.GroupChatId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.DeleteGroupChat(&groupChatId, executorId)
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := DeleteGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// RenameGroupChat is a handler for renaming a group chat.
func (g *GroupChatController) RenameGroupChat(c *gin.Context) {
	var jsonRequestBody RenameGroupChatRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	slog.Info("RenameGroupChat", "jsonRequestBody", jsonRequestBody)

	groupChatId, err := validator.ValidateGroupChatId(jsonRequestBody.GroupChatId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatName, err := validator.ValidateGroupChatName(jsonRequestBody.Name).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.RenameGroupChat(&groupChatId, groupChatName, executorId)
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := RenameGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// AddMember is a handler for adding a member to a group chat.
func (g *GroupChatController) AddMember(c *gin.Context) {
	var jsonRequestBody AddMemberRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatId, err := validator.ValidateGroupChatId(jsonRequestBody.GroupChatId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	accountId, err := validator.ValidateUserAccountId(jsonRequestBody.UserAccountId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	role, err := models.StringToRole(jsonRequestBody.Role)
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.AddMember(&groupChatId, accountId, role, executorId)

	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := AddMemberResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// RemoveMember is a handler for removing a member from a group chat.
func (g *GroupChatController) RemoveMember(c *gin.Context) {
	var jsonRequestBody RemoveMemberRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatId, err := validator.ValidateGroupChatId(jsonRequestBody.GroupChatId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	userAccountId, err := validator.ValidateUserAccountId(jsonRequestBody.UserAccountId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.RemoveMember(&groupChatId, userAccountId, executorId)
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := RemoveMemberResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// PostMessage is a handler for posting a message to a group chat.
func (g *GroupChatController) PostMessage(c *gin.Context) {
	var jsonRequestBody PostMessageRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatId, err := validator.ValidateGroupChatId(jsonRequestBody.GroupChatId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	messageId := models.NewMessageId()

	senderId, err := validator.ValidateUserAccountId(jsonRequestBody.UserAccountId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	message, err := validator.ValidateMessage(messageId, jsonRequestBody.Message, senderId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.PostMessage(&groupChatId, message, executorId)
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := PostMessageResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString(), MessageId: messageId.String()}
	c.JSON(http.StatusOK, response)
}

// DeleteMessage is a handler for deleting a message from a group chat.
func (g *GroupChatController) DeleteMessage(c *gin.Context) {
	var jsonRequestBody DeleteMessageRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	groupChatId, err := validator.ValidateGroupChatId(jsonRequestBody.GroupChatId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	messageId, err := validator.ValidateMessageId(jsonRequestBody.MessageId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		handleClientError(c, http.StatusBadRequest, err)
		return
	}

	event, err := g.groupChatCommandProcessor.DeleteMessage(&groupChatId, messageId, executorId)
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := DeleteMessageResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// handleClientError is a helper function for handling client errors.
func handleClientError(c *gin.Context, statusCode int, err error) {
	response := GroupChatResponseErrorBody{Message: err.Error()}
	c.JSON(statusCode, response)
}
