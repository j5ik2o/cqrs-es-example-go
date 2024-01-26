package api

import (
	"bytes"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
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
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	actualGroupChat, err := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet()).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
}

func Test_GroupChat_Delete(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	err = sender.sendDeleteGroupChatCommand(recorder, groupChatID, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	actualGroupChat, err := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet()).Get()
	require.NoError(t, err)
	require.True(t, actualGroupChat.IsDeleted())
}

func Test_GroupChat_Rename(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	groupChatName2 := "test2"
	err = sender.sendRenameGroupChatCommand(recorder, groupChatID, groupChatName2, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	actualGroupChat, err := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet()).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName2, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
}

func Test_GroupChat_AddMember(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountId := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatID, userAccountId, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	actualGroupChat, err := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet()).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	require.True(t, actualGroupChat.GetMembers().FindByUserAccountId(models.NewUserAccountIdFromString(userAccountId).MustGet()).IsPresent())
}

func Test_GroupChat_RemoveMember(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountId := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatID, userAccountId, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	err = sender.sendRemoveMemberCommand(recorder, groupChatID, userAccountId, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	result := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet())
	require.True(t, result.IsOk())
	actualGroupChat := result.MustGet()
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	require.False(t, actualGroupChat.GetMembers().FindByUserAccountId(models.NewUserAccountIdFromString(userAccountId).MustGet()).IsPresent())
}

func Test_GroupChat_PostMessage(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountId := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatID, userAccountId, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	message := "test-message"
	err = sender.sendPostMessageCommand(recorder, groupChatID, userAccountId, message, userAccountId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	actualGroupChat, err := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet()).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	require.True(t, actualGroupChat.GetMembers().FindByUserAccountId(models.NewUserAccountIdFromString(userAccountId).MustGet()).IsPresent())
	require.Equal(t, message, actualGroupChat.GetMessages().ToSlice()[0].GetText())
}

func Test_GroupChat_DeleteMessage(t *testing.T) {
	groupChatRepository := repository.NewGroupChatRepository(eventstoreadaptergo.NewEventStoreOnMemory())
	groupChatController := NewGroupChatController(groupChatRepository)

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

	groupChatID, err := getGroupChatId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	userAccountId := "01HMGVNJTTW24CHABMT85M9EN9"
	err = sender.sendAddMemberCommand(recorder, groupChatID, userAccountId, executorId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	message := "test-message"
	err = sender.sendPostMessageCommand(recorder, groupChatID, userAccountId, message, userAccountId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	messageId, err := getMessageId(recorder)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	err = sender.sendDeleteMessageCommand(recorder, groupChatID, messageId, userAccountId)
	require.NoError(t, err)
	require.Equal(t, 200, recorder.Code)

	actualGroupChat, err := groupChatRepository.FindById(models.NewGroupChatIdFromString(groupChatID).MustGet()).Get()
	require.NoError(t, err)
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
	require.Equal(t, groupChatName, actualGroupChat.GetName().String())
	require.Equal(t, executorId, actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId().GetValue())
	require.True(t, actualGroupChat.GetMembers().FindByUserAccountId(models.NewUserAccountIdFromString(userAccountId).MustGet()).IsPresent())
	require.False(t, actualGroupChat.GetMessages().Contains(models.NewMessageIdFromString(messageId).MustGet()))
}

type RequestSender struct {
	engine *gin.Engine
}

func NewRequestSender(engine *gin.Engine) *RequestSender {
	return &RequestSender{
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

func (r *RequestSender) sendAddMemberCommand(w *httptest.ResponseRecorder, groupChatID string, accountId string, executorId string) error {
	requestBodyJson, err := json.Marshal(AddMemberRequestBody{
		GroupChatId: groupChatID,
		AccountId:   accountId,
		Role:        "admin",
		ExecutorId:  executorId,
	})
	if err != nil {
		return err
	}
	requestBodyJsonBody := bytes.NewBuffer(requestBodyJson)
	request := httptest.NewRequest("POST", "/group-chats/add-member", requestBodyJsonBody)
	r.engine.ServeHTTP(w, request)
	return nil
}

func (r *RequestSender) sendRemoveMemberCommand(w *httptest.ResponseRecorder, groupChatID string, accountId string, executorId string) error {
	requestBodyJson, err := json.Marshal(RemoveMemberRequestBody{
		GroupChatId: groupChatID,
		AccountId:   accountId,
		ExecutorId:  executorId,
	})
	if err != nil {
		return err
	}
	requestBodyJsonBody := bytes.NewBuffer(requestBodyJson)
	request := httptest.NewRequest("POST", "/group-chats/remove-member", requestBodyJsonBody)
	r.engine.ServeHTTP(w, request)
	return nil
}

func (r *RequestSender) sendPostMessageCommand(w *httptest.ResponseRecorder, groupChatID string, accountId string, message string, executorId string) error {
	requestBodyJson, err := json.Marshal(PostMessageRequestBody{
		GroupChatId: groupChatID,
		AccountId:   accountId,
		Message:     message,
		ExecutorId:  executorId,
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
