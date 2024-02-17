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
// @Summary Create a group chat
// @Description Create a group chat
// @Param   body body CreateGroupChatRequestBody true "CreateGroupChatRequestBody"
// @Success 200 {object} CreateGroupChatResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/create [post]
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

	event, err := g.groupChatCommandProcessor.CreateGroupChat(groupChatName, executorId).Get()
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := CreateGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// DeleteGroupChat is a handler for deleting a group chat.
// @Summary Delete a group chat
// @Description Delete a group chat
// @Param   body body DeleteGroupChatRequestBody true "DeleteGroupChatRequestBody"
// @Success 200 {object} DeleteGroupChatResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/delete [post]
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

	event, err := g.groupChatCommandProcessor.DeleteGroupChat(&groupChatId, executorId).Get()
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := DeleteGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// RenameGroupChat is a handler for renaming a group chat.
// @Summary Rename a group chat
// @Description Rename a group chat
// @Param   body body RenameGroupChatRequestBody true "RenameGroupChatRequestBody"
// @Success 200 {object} RenameGroupChatResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/rename [post]
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

	event, err := g.groupChatCommandProcessor.RenameGroupChat(&groupChatId, groupChatName, executorId).Get()
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := RenameGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// AddMember is a handler for adding a member to a group chat.
// @Summary Add a member to a group chat
// @Description Add a member to a group chat
// @Param   body body AddMemberRequestBody true "AddMemberRequestBody"
// @Success 200 {object} AddMemberResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/add-member [post]
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

	event, err := g.groupChatCommandProcessor.AddMember(&groupChatId, accountId, role, executorId).Get()

	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := AddMemberResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// RemoveMember is a handler for removing a member from a group chat.
// @Summary Remove a member from a group chat
// @Description Remove a member from a group chat
// @Param   body body RemoveMemberRequestBody true "RemoveMemberRequestBody"
// @Success 200 {object} RemoveMemberResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/remove-member [post]
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

	event, err := g.groupChatCommandProcessor.RemoveMember(&groupChatId, userAccountId, executorId).Get()
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := RemoveMemberResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}

// PostMessage is a handler for posting a message to a group chat.
// @Summary Post a message to a group chat
// @Description Post a message to a group chat
// @Param   body body PostMessageRequestBody true "PostMessageRequestBody"
// @Success 200 {object} PostMessageResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/post-message [post]
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

	event, err := g.groupChatCommandProcessor.PostMessage(&groupChatId, message, executorId).Get()
	if err != nil {
		response := GroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := PostMessageResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString(), MessageId: messageId.String()}
	c.JSON(http.StatusOK, response)
}

// DeleteMessage is a handler for deleting a message from a group chat.
// @Summary Delete a message from a group chat
// @Description Delete a message from a group chat
// @Param   body body DeleteMessageRequestBody true "DeleteMessageRequestBody"
// @Success 200 {object} DeleteMessageResponseSuccessBody
// @Failure 400 {object} GroupChatResponseErrorBody
// @Failure 500 {object} GroupChatResponseErrorBody
// @Router /v1/group-chat/delete-message [post]
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

	event, err := g.groupChatCommandProcessor.DeleteMessage(&groupChatId, messageId, executorId).Get()
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
