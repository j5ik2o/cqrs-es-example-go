package domain

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_GroupChat_AddMember(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId)
	memberId := models.NewMemberId()
	userAccountId := models.NewUserAccountId()

	// When
	tuple2, err := groupChat.AddMember(memberId, userAccountId, models.MemberRole, adminId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple2.V1.seqNr)
	require.True(t, tuple2.V1.GetMembers().FindByUserAccountId(&userAccountId).IsPresent())
	require.Equal(t, &groupChat.id, tuple2.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple2.V2.GetSeqNr())
}

func Test_GroupChat_RemoveMemberByUserAccountId(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId)
	memberId := models.NewMemberId()
	userAccountId := models.NewUserAccountId()
	result := groupChat.AddMember(memberId, userAccountId, models.MemberRole, adminId)
	require.True(t, result.IsOk())
	groupChat = result.MustGet().V1

	// When
	tuple2, err := groupChat.RemoveMemberByUserAccountId(userAccountId, adminId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple2.V1.seqNr)
	require.False(t, tuple2.V1.GetMembers().FindByUserAccountId(&userAccountId).IsPresent())
	require.Equal(t, &groupChat.id, tuple2.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple2.V2.GetSeqNr())
}

func Test_GroupChat_Rename(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId)
	name := models.NewGroupChatName("test2").MustGet()

	// When
	tuple2, err := groupChat.Rename(name, adminId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple2.V1.seqNr)
	require.Equal(t, "test2", tuple2.V1.GetName().String())
	require.Equal(t, &groupChat.id, tuple2.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple2.V2.GetSeqNr())
}

func Test_GroupChat_Delete(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId)

	// When
	tuple2, err := groupChat.Delete(adminId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple2.V1.seqNr)
	require.True(t, tuple2.V1.IsDeleted())
	require.Equal(t, &groupChat.id, tuple2.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple2.V2.GetSeqNr())
}

func Test_GroupChat_PostMessage(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	userAccountId := models.NewUserAccountId()
	groupChat, _ := NewGroupChat(groupChatName, adminId)
	groupChat = groupChat.AddMember(models.NewMemberId(), userAccountId, models.MemberRole, adminId).MustGet().V1
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", userAccountId).MustGet()

	// When
	tuple1, err := groupChat.PostMessage(message, userAccountId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple1.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple1.V1.seqNr)
	require.True(t, tuple1.V1.GetMessages().Get(&messageId).IsPresent())
	require.Equal(t, message, tuple1.V1.GetMessages().Get(&messageId).MustGet())
}

func Test_GroupChat_EditMessage(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	userAccountId := models.NewUserAccountId()
	groupChat, _ := NewGroupChat(groupChatName, adminId)
	groupChat = groupChat.AddMember(models.NewMemberId(), userAccountId, models.MemberRole, adminId).MustGet().V1
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", userAccountId).MustGet()

	tuple1, err := groupChat.PostMessage(message, userAccountId).Get()
	groupChat = tuple1.V1

	message = message.WithText("test2").MustGet()
	tuple2, err := groupChat.EditMessage(message, userAccountId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple2.V1.seqNr)
	require.True(t, tuple2.V1.GetMessages().Get(&messageId).IsPresent())
	require.Equal(t, message, tuple2.V1.GetMessages().Get(&messageId).MustGet())
}

func Test_GroupChat_DeleteMessage(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	userAccountId := models.NewUserAccountId()
	groupChat, _ := NewGroupChat(groupChatName, adminId)
	groupChat = groupChat.AddMember(models.NewMemberId(), userAccountId, models.MemberRole, adminId).MustGet().V1
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", userAccountId).MustGet()
	result1 := groupChat.PostMessage(message, userAccountId)
	tuple1 := result1.MustGet()

	// When
	tuple2, err := tuple1.V1.DeleteMessage(messageId, userAccountId).Get()

	// Then
	require.NoError(t, err)
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+2, tuple2.V1.seqNr)
	require.True(t, tuple2.V1.GetMessages().Get(&messageId).IsAbsent())
}
