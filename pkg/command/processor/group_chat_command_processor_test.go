package processor

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	esa "github.com/j5ik2o/event-store-adapter-go/v5"
	"github.com/j5ik2o/event-store-adapter-go/v5/memory"
	"github.com/stretchr/testify/require"
	"testing"
)

func newMemoryEventStore(t *testing.T) esa.EventStore[events.GroupChatEvent, domain.GroupChat] {
	t.Helper()
	store, err := memory.NewStore()
	require.NoError(t, err)
	eventStore, err := memory.New(store, repository.NewEventSerializer(), repository.NewSnapshotSerializer())
	require.NoError(t, err)
	return eventStore
}

func Test_CreateGroupChat(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	// When
	result, err := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	// Then
	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatCreated)
	require.True(t, ok)
	require.True(t, groupName.Equals(event.GetName()))
	require.Equal(t, uint64(1), result.V1.GetSeqNr())
	require.True(t, result.V1.GetGroupChatId().GetValue() != "")
}

func Test_DeleteGroupChat(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	// When
	result, err = commandProcessor.DeleteGroupChat(groupChatId, executorId).Get()
	// Then
	require.NoError(t, err)
	_, ok := result.V2.(*events.GroupChatDeleted)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
}

func Test_RenameGroupChat(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	newGroupName := models.NewGroupChatName("test2").MustGet()
	// When
	result, err = commandProcessor.RenameGroupChat(groupChatId, newGroupName, executorId).Get()
	// Then
	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatRenamed)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
	require.True(t, newGroupName.Equals(event.GetName()))
}

func Test_AddMember(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	// When
	result, err = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId).Get()
	// Then
	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatMemberAdded)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
	require.Equal(t, memberRole, event.GetMember().GetRole())
	require.True(t, memberUserAccountId.Equals(event.GetMember().GetUserAccountId()))
}

func Test_RemoveMember(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId).Get()
	// When
	result, err = commandProcessor.RemoveMember(groupChatId, memberUserAccountId, executorId).Get()
	// Then
	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatMemberRemoved)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
	require.True(t, memberUserAccountId.Equals(event.GetUserAccountId()))
}

func Test_PostMessage(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId).Get()
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", memberUserAccountId).MustGet()
	// When
	result, err = commandProcessor.PostMessage(groupChatId, message, memberUserAccountId).Get()
	// Then
	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatMessagePosted)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
	require.True(t, message.Equals(event.GetMessage()))
}

func Test_EditMessage(t *testing.T) {
	// Give
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId).Get()
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", memberUserAccountId).MustGet()
	_, _ = commandProcessor.PostMessage(groupChatId, message, memberUserAccountId).Get()
	message = message.WithText("test2").MustGet()
	result, err = commandProcessor.EditMessage(groupChatId, message, memberUserAccountId).Get()

	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatMessageEdited)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
	require.True(t, message.Equals(event.GetMessage()))
}

func Test_DeleteMessage(t *testing.T) {
	// Given
	groupChatRepository, err := repository.NewGroupChatRepository(newMemoryEventStore(t))
	require.NoError(t, err)
	commandProcessor := NewGroupChatCommandProcessor(&groupChatRepository)
	groupName := models.NewGroupChatName("test").MustGet()
	executorId := models.NewUserAccountId()
	result, _ := commandProcessor.CreateGroupChat(groupName, executorId).Get()
	groupChatId := result.V1.GetGroupChatId()
	memberUserAccountId := models.NewUserAccountId()
	var memberRole models.Role = models.MemberRole
	_, _ = commandProcessor.AddMember(groupChatId, memberUserAccountId, memberRole, executorId).Get()
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", memberUserAccountId).MustGet()
	_, _ = commandProcessor.PostMessage(groupChatId, message, memberUserAccountId).Get()
	// When
	result, err = commandProcessor.DeleteMessage(groupChatId, messageId, memberUserAccountId).Get()
	// Then
	require.NoError(t, err)
	event, ok := result.V2.(*events.GroupChatMessageDeleted)
	require.True(t, ok)
	require.True(t, groupChatId.Equals(result.V1.GetGroupChatId()))
	require.True(t, messageId.Equals(event.GetMessageId()))
}
