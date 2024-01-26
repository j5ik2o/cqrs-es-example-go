package useCase

import (
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	event_store_adapter_go "github.com/j5ik2o/event-store-adapter-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CreateGroupChat(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	// When
	result, err := commandProcessor.CreateGroupChat(groupName, executorId)
	// Then
	require.NoError(t, err)
	require.True(t, result.IsCreated())
	event, ok := result.(*events.GroupChatCreated)
	require.True(t, ok)
	_, ok = event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.Equal(t, groupName, event.GetName())
}

func Test_DeleteGroupChat(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId)
	groupChatId, _ := result.GetAggregateId().(*models.GroupChatId)
	// When
	result, err := commandProcessor.DeleteGroupChat(groupChatId, executorId)
	// Then
	require.NoError(t, err)
	event, ok := result.(*events.GroupChatDeleted)
	require.True(t, ok)
	actualGroupChatId, ok := event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(actualGroupChatId))
}

func Test_RenameGroupChat(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId)
	groupChatId, _ := result.GetAggregateId().(*models.GroupChatId)
	newGroupName := models.NewGroupChatName("test2").MustGet()
	// When
	result, err := commandProcessor.RenameGroupChat(groupChatId, newGroupName, executorId)
	// Then
	require.NoError(t, err)
	event, ok := result.(*events.GroupChatRenamed)
	require.True(t, ok)
	actualGroupChatId, ok := event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(actualGroupChatId))
	require.Equal(t, newGroupName, event.GetName())
}

func Test_AddMember(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId)
	groupChatId, _ := result.GetAggregateId().(*models.GroupChatId)
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	// When
	result, err := commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId)
	// Then
	require.NoError(t, err)
	event, ok := result.(*events.GroupChatMemberAdded)
	require.True(t, ok)
	actualGroupChatId, ok := event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(actualGroupChatId))
	require.Equal(t, memberRole, event.GetMember().GetRole())
	require.True(t, memberUserAccountId.Equals(event.GetMember().GetUserAccountId()))
}

func Test_RemoveMember(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId)
	groupChatId, _ := result.GetAggregateId().(*models.GroupChatId)
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId)
	// When
	result, err := commandProcessor.RemoveMember(groupChatId, memberUserAccountId, executorId)
	// Then
	require.NoError(t, err)
	event, ok := result.(*events.GroupChatMemberRemoved)
	require.True(t, ok)
	actualGroupChatId, ok := event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(actualGroupChatId))
	require.True(t, memberUserAccountId.Equals(event.GetUserAccountId()))
}

func Test_PostMessage(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId)
	groupChatId, _ := result.GetAggregateId().(*models.GroupChatId)
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId)
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", memberUserAccountId).MustGet()
	// When
	result, err := commandProcessor.PostMessage(groupChatId, message, memberUserAccountId)
	// Then
	require.NoError(t, err)
	event, ok := result.(*events.GroupChatMessagePosted)
	require.True(t, ok)
	actualGroupChatId, ok := event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(actualGroupChatId))
	require.True(t, message.Equals(event.GetMessage()))
}

func Test_DeleteMessage(t *testing.T) {
	// Given
	groupChatRepository := repository.NewGroupChatRepository(event_store_adapter_go.NewEventStoreOnMemory())
	commandProcessor := NewGroupChatCommandProcessor(groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId)
	groupChatId, _ := result.GetAggregateId().(*models.GroupChatId)
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId)
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", memberUserAccountId).MustGet()
	_, _ = commandProcessor.PostMessage(groupChatId, message, memberUserAccountId)
	// When
	result, err := commandProcessor.DeleteMessage(groupChatId, messageId, memberUserAccountId)
	// Then
	require.NoError(t, err)
	event, ok := result.(*events.GroupChatMessageDeleted)
	require.True(t, ok)
	actualGroupChatId, ok := event.GetAggregateId().(*models.GroupChatId)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(actualGroupChatId))
	require.True(t, messageId.Equals(event.GetMessageId()))
}
