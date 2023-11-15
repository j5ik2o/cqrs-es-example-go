package api

import (
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/validator"
	"cqrs-es-example-go/pkg/command/useCase"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CreateGroupChatRequestBody struct {
	Name       string `json:"name"`
	ExecutorId string `json:"executorId"`
}

type CreateGroupChatResponseSuccessBody struct {
	GroupChatId string `json:"groupChatId"`
}

type CreateGroupChatResponseErrorBody struct {
	Message string `json:"message"`
}

type GroupChatController struct {
	repository repository.GroupChatRepository
}

func NewGroupChatController(repository repository.GroupChatRepository) *GroupChatController {
	return &GroupChatController{
		repository,
	}
}

func (g *GroupChatController) CreateGroupChat(c *gin.Context) {
	var jsonRequestBody CreateGroupChatRequestBody

	if err := c.ShouldBindJSON(&jsonRequestBody); err != nil {
		response := CreateGroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	groupChatName, err := validator.ValidateGroupChatName(jsonRequestBody.Name).Get()
	if err != nil {
		response := CreateGroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	executorId, err := validator.ValidateUserAccountId(jsonRequestBody.ExecutorId).Get()
	if err != nil {
		response := CreateGroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	commandProcessor := useCase.NewGroupChatCommandProcessor(g.repository)
	event, err := commandProcessor.CreateGroupChat(groupChatName, executorId, executorId)

	if err != nil {
		response := CreateGroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := CreateGroupChatResponseSuccessBody{GroupChatId: event.GetAggregateId().AsString()}
	c.JSON(http.StatusOK, response)
}
