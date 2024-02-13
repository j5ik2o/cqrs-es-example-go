package api

import (
	"bytes"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"cqrs-es-example-go/pkg/command/useCase"
	"encoding/json"
	"github.com/gin-gonic/gin"
	eventstoreadaptergo "github.com/j5ik2o/event-store-adapter-go"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"testing"
)

func Test_GroupChat_Create(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
	}
	sender := NewRequestSender(engine)

	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatIdString, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
}

func Test_GroupChat_Delete(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
		groupChat.POST("/delete", groupChatController.DeleteGroupChat)
	}
	sender := NewRequestSender(engine)

	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	err = sender.sendDeleteGroupChatCommand(recorder, groupChatIdString, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.True(t, actualGroupChat.IsDeleted())
}

func Test_GroupChat_Rename(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
		groupChat.POST("/rename", groupChatController.RenameGroupChat)
	}
	sender := NewRequestSender(engine)

	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	groupChatName2 := "test2"
	err = sender.sendRenameGroupChatCommand(recorder, groupChatIdString, groupChatName2, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatIdString, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName2, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
}

func Test_GroupChat_AddMember(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
		groupChat.POST("/add-member", groupChatController.AddMember)
	}

	sender := NewRequestSender(engine)
	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountIdString := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatIdString, userAccountIdString, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatIdString, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	userAccountId := models.NewUserAccountIdFromString(userAccountIdString).MustGet()
	require.True(t, actualGroupChat.GetMembers().FindByUserAccountId(&userAccountId).IsPresent())
}

func Test_GroupChat_RemoveMember(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
		groupChat.POST("/add-member", groupChatController.AddMember)
		groupChat.POST("/remove-member", groupChatController.RemoveMember)
	}

	sender := NewRequestSender(engine)
	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountIdString := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatIdString, userAccountIdString, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	err = sender.sendRemoveMemberCommand(recorder, groupChatIdString, userAccountIdString, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatIdString, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	userAccountId := models.NewUserAccountIdFromString(userAccountIdString).MustGet()
	require.False(t, actualGroupChat.GetMembers().FindByUserAccountId(&userAccountId).IsPresent())
}

func Test_GroupChat_PostMessage(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
		groupChat.POST("/add-member", groupChatController.AddMember)
		groupChat.POST("/post-message", groupChatController.PostMessage)
	}

	sender := NewRequestSender(engine)
	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountIdString := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatIdString, userAccountIdString, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	message := "test-message"
	err = sender.sendPostMessageCommand(recorder, groupChatIdString, userAccountIdString, message, userAccountIdString)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatIdString, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	userAccountId := models.NewUserAccountIdFromString(userAccountIdString).MustGet()
	require.True(t, actualGroupChat.GetMembers().FindByUserAccountId(&userAccountId).IsPresent())
	require.Equal(t, message, actualGroupChat.GetMessages().ToSlice()[0].GetText())
}

func Test_GroupChat_DeleteMessage(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatCommandProcessor := useCase.NewGroupChatCommandProcessor(&groupChatRepository)
	groupChatController := NewGroupChatController(groupChatCommandProcessor)

	engine := gin.Default()
	groupChat := engine.Group("/group-chats")
	{
		groupChat.POST("/create", groupChatController.CreateGroupChat)
		groupChat.POST("/add-member", groupChatController.AddMember)
		groupChat.POST("/post-message", groupChatController.PostMessage)
		groupChat.POST("/delete-message", groupChatController.DeleteMessage)
	}

	sender := NewRequestSender(engine)
	recorder := httptest.NewRecorder()
	groupChatName := "test1"
	executorId := "01H42K4ABWQ5V2XQEP3A48VE0Z"
	err := sender.sendCreateGroupChatCommand(recorder, groupChatName, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatIdString, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountIdString := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatIdString, userAccountIdString, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	message := "test-message"
	err = sender.sendPostMessageCommand(recorder, groupChatIdString, userAccountIdString, message, userAccountIdString)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	messageIdString, err := getMessageId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	err = sender.sendDeleteMessageCommand(recorder, groupChatIdString, messageIdString, userAccountIdString)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	groupChatId := models.NewGroupChatIdFromString(groupChatIdString).MustGet()
	actualGroupChat, err := groupChatRepository.FindById(&groupChatId).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatIdString, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	userAccountId := models.NewUserAccountIdFromString(userAccountIdString).MustGet()
	require.True(t, actualGroupChat.GetMembers().FindByUserAccountId(&userAccountId).IsPresent())
	messageId := models.NewMessageIdFromString(messageIdString).MustGet()
	require.False(t, actualGroupChat.GetMessages().Contains(&messageId))
}

type RequestSender struct {
	engine *gin.Engine
}

func NewRequestSender(engine *gin.Engine) RequestSender {
	return RequestSender{
		engine: engine,
	}
}

func (r *RequestSender) sendCreateGroupChatCommand(w *httptest.ResponseRecorder, groupChatName string, executorId string) error {
	createGroupChatRequestBodyJson, err := json.Marshal(CreateGroupChatRequestBody{
		Name:       groupChatName,
		ExecutorId: executorId,
	})
	if err != nil {
		return err
	}
	createGroupChatRequestBody := bytes.NewBuffer(createGroupChatRequestBodyJson)
	createGroupChatRequest := httptest.NewRequest("POST", "/group-chats/create", createGroupChatRequestBody)
	r.engine.ServeHTTP(w, createGroupChatRequest)
	return nil
}

func (r *RequestSender) sendDeleteGroupChatCommand(w *httptest.ResponseRecorder, groupChatID string, executorId string) error {
	createGroupChatRequestBodyJson, err := json.Marshal(DeleteGroupChatRequestBody{
		GroupChatId: groupChatID,
		ExecutorId:  executorId,
	})
	if err != nil {
		return err
	}
	createGroupChatRequestBody := bytes.NewBuffer(createGroupChatRequestBodyJson)
	createGroupChatRequest := httptest.NewRequest("POST", "/group-chats/delete", createGroupChatRequestBody)
	r.engine.ServeHTTP(w, createGroupChatRequest)
	return nil
}

func getGroupChatId(w *httptest.ResponseRecorder) (string, error) {
	responseBody, err := io.ReadAll(w.Body)
	if err != nil {
		return "", err
	}
	var createGroupChatResponseSuccessBody CreateGroupChatResponseSuccessBody
	err = json.Unmarshal(responseBody, &createGroupChatResponseSuccessBody)
	if err != nil {
		return "", err
	}
	groupChatID := createGroupChatResponseSuccessBody.GroupChatId
	return groupChatID, err
}

func getMessageId(w *httptest.ResponseRecorder) (string, error) {
	responseBody, err := io.ReadAll(w.Body)
	if err != nil {
		return "", err
	}
	var responseSuccessBody PostMessageResponseSuccessBody
	err = json.Unmarshal(responseBody, &responseSuccessBody)
	if err != nil {
		return "", err
	}
	messageId := responseSuccessBody.MessageId
	return messageId, err
}

func (r *RequestSender) sendRenameGroupChatCommand(w *httptest.ResponseRecorder, groupChatID string, groupChatName string, executorId string) error {
	renameGroupChatRequestBodyJson, err := json.Marshal(RenameGroupChatRequestBody{
		GroupChatId: groupChatID,
		Name:        groupChatName,
		ExecutorId:  executorId,
	})
	if err != nil {
		return err
	}
	renameGroupChatRequestBodyJsonBody := bytes.NewBuffer(renameGroupChatRequestBodyJson)
	renameGroupChatRequest := httptest.NewRequest("POST", "/group-chats/rename", renameGroupChatRequestBodyJsonBody)
	r.engine.ServeHTTP(w, renameGroupChatRequest)
	return nil
}

func (r *RequestSender) sendAddMemberCommand(w *httptest.ResponseRecorder, groupChatID string, userAccountId string, executorId string) error {
	requestBodyJson, err := json.Marshal(AddMemberRequestBody{
		GroupChatId:   groupChatID,
		UserAccountId: userAccountId,
		Role:          models.Role(models.AdminRole).String(),
		ExecutorId:    executorId,
	})
	if err != nil {
		return err
	}
	requestBodyJsonBody := bytes.NewBuffer(requestBodyJson)
	request := httptest.NewRequest("POST", "/group-chats/add-member", requestBodyJsonBody)
	r.engine.ServeHTTP(w, request)
	return nil
}

func (r *RequestSender) sendRemoveMemberCommand(w *httptest.ResponseRecorder, groupChatID string, userAccountId string, executorId string) error {
	requestBodyJson, err := json.Marshal(RemoveMemberRequestBody{
		GroupChatId:   groupChatID,
		UserAccountId: userAccountId,
		ExecutorId:    executorId,
	})
	if err != nil {
		return err
	}
	requestBodyJsonBody := bytes.NewBuffer(requestBodyJson)
	request := httptest.NewRequest("POST", "/group-chats/remove-member", requestBodyJsonBody)
	r.engine.ServeHTTP(w, request)
	return nil
}

func (r *RequestSender) sendPostMessageCommand(w *httptest.ResponseRecorder, groupChatID string, userAccountId string, message string, executorId string) error {
	requestBodyJson, err := json.Marshal(PostMessageRequestBody{
		GroupChatId:   groupChatID,
		UserAccountId: userAccountId,
		Message:       message,
		ExecutorId:    executorId,
	})
	if err != nil {
		return err
	}
	requestBodyJsonBody := bytes.NewBuffer(requestBodyJson)
	request := httptest.NewRequest("POST", "/group-chats/post-message", requestBodyJsonBody)
	r.engine.ServeHTTP(w, request)
	return nil
}

func (r *RequestSender) sendDeleteMessageCommand(w *httptest.ResponseRecorder, groupChatID string, messageId string, executorId string) error {
	requestBodyJson, err := json.Marshal(DeleteMessageRequestBody{
		GroupChatId: groupChatID,
		MessageId:   messageId,
		ExecutorId:  executorId,
	})
	if err != nil {
		return err
	}
	requestBodyJsonBody := bytes.NewBuffer(requestBodyJson)
	request := httptest.NewRequest("POST", "/group-chats/delete-message", requestBodyJsonBody)
	r.engine.ServeHTTP(w, request)
	return nil
}
