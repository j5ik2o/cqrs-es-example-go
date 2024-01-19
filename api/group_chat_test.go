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

	w1 := httptest.NewRecorder()
	err := sendCreateGroupChatCommand("test1", "01H42K4ABWQ5V2XQEP3A48VE0Z", engine, w1)
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

	w1 := httptest.NewRecorder()
	err := sendCreateGroupChatCommand("test1", "01H42K4ABWQ5V2XQEP3A48VE0Z", engine, w1)
	require.NoError(t, err)
	require.Equal(t, 200, w1.Code)

	groupChatID, err := getGroupChatId(w1)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	err = sendRenameGroupChatCommand(groupChatID, "test2", "01H42K4ABWQ5V2XQEP3A48VE0Z", engine, w2)
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

	w1 := httptest.NewRecorder()
	err := sendCreateGroupChatCommand("test1", "01H42K4ABWQ5V2XQEP3A48VE0Z", engine, w1)
	require.NoError(t, err)
	require.Equal(t, 200, w1.Code)

	groupChatID, err := getGroupChatId(w1)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	err = sendAddMemberCommand(groupChatID, engine, w2)
	require.NoError(t, err)
	require.Equal(t, 200, w2.Code)
}

func sendCreateGroupChatCommand(groupChatName string, executorId string, engine *gin.Engine, w *httptest.ResponseRecorder) error {
	createGroupChatRequestBodyJson, err := json.Marshal(CreateGroupChatRequestBody{
		Name:       groupChatName,
		ExecutorId: executorId,
	})
	if err != nil {
		return err
	}
	createGroupChatRequestBody := bytes.NewBuffer(createGroupChatRequestBodyJson)
	createGroupChatRequest := httptest.NewRequest("POST", "/group-chats/create", createGroupChatRequestBody)
	engine.ServeHTTP(w, createGroupChatRequest)
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

func sendRenameGroupChatCommand(groupChatID string, groupChatName string, executorId string, engine *gin.Engine, w *httptest.ResponseRecorder) error {
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
	engine.ServeHTTP(w, renameGroupChatRequest)
	return nil
}

func sendAddMemberCommand(groupChatID string, engine *gin.Engine, w *httptest.ResponseRecorder) error {
	addMemberRequestBodyJson, err := json.Marshal(AddMemberRequestBody{
		GroupChatId: groupChatID,
		AccountId:   "1",
		Role:        "admin",
		ExecutorId:  "01H42K4ABWQ5V2XQEP3A48VE0Z",
	})
	if err != nil {
		return err
	}
	addMemberRequestBodyJsonBody := bytes.NewBuffer(addMemberRequestBodyJson)
	addMemberRequest := httptest.NewRequest("POST", "/group-chats/add-member", addMemberRequestBodyJsonBody)
	engine.ServeHTTP(w, addMemberRequest)
	return nil
}
