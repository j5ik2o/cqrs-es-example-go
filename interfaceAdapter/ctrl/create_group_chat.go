package ctrl

import (
	"cqrs-es-example-go/repository"
	"cqrs-es-example-go/useCase"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupChatName, err := ValidateGroupChatName(jsonRequestBody.Name)
	if err != nil {
		response := CreateGroupChatResponseErrorBody{Message: err.Error()}
		c.JSON(http.StatusBadRequest, response)
		return
	}
	executorId, err := ValidateUserAccountId(jsonRequestBody.ExecutorId)
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
