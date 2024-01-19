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

	w1 := httptest.NewRecorder()
	err := sender.sendCreateGroupChatCommand(w1, "test1", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w1.Code)

	groupChatID, err := getGroupChatId(w1)
	require.NoError(t, err)
	id := models.NewGroupChatIdFromString(groupChatID).MustGet()
	result := groupChatRepository.FindById(id)
	require.True(t, result.IsOk())
	actualGroupChat := result.MustGet()
	require.Equal(t, groupChatID, actualGroupChat.GetId().String())
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

	w1 := httptest.NewRecorder()
	err := sender.sendCreateGroupChatCommand(w1, "test1", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w1.Code)

	groupChatID, err := getGroupChatId(w1)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	err = sender.sendRenameGroupChatCommand(w2, groupChatID, "test2", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w2.Code)
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
	w1 := httptest.NewRecorder()
	err := sender.sendCreateGroupChatCommand(w1, "test1", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w1.Code)

	groupChatID, err := getGroupChatId(w1)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	err = sender.sendAddMemberCommand(w2, groupChatID, "01H42K4ABWQ5V2XQEP3A48VE0Z", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w2.Code)
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
	w1 := httptest.NewRecorder()
	err := sender.sendCreateGroupChatCommand(w1, "test1", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w1.Code)

	groupChatID, err := getGroupChatId(w1)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	err = sender.sendAddMemberCommand(w2, groupChatID, "01H42K4ABWQ5V2XQEP3A48VE0Z", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w2.Code)

	w3 := httptest.NewRecorder()
	err = sender.sendRemoveMemberCommand(w3, groupChatID, "01H42K4ABWQ5V2XQEP3A48VE0Z", "01H42K4ABWQ5V2XQEP3A48VE0Z")
	require.NoError(t, err)
	require.Equal(t, 200, w3.Code)
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
