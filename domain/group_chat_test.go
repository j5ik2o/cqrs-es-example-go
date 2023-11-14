package domain

import (
	"cqrs-es-example-go/domain/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGroupChat_AddMember(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId, adminId)
	memberId := models.NewMemberId()
	userAccountId := models.NewUserAccountId()

	// When
	result := groupChat.AddMember(memberId, userAccountId, models.MemberRole, adminId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()

	require.Equal(t, groupChat.id, tuple.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple.V1.seqNr)
	require.True(t, tuple.V1.GetMembers().FindByUserAccountId(userAccountId).IsPresent())

	require.Equal(t, groupChat.id, tuple.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple.V2.GetSeqNr())
}

func TestGroupChat_RemoveMemberByUserAccountId(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId, adminId)
	memberId := models.NewMemberId()
	userAccountId := models.NewUserAccountId()
	result := groupChat.AddMember(memberId, userAccountId, models.MemberRole, adminId)
	require.True(t, result.IsOk())

	// When
	result = groupChat.RemoveMemberByUserAccountId(userAccountId, adminId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()
	require.Equal(t, groupChat.id, tuple.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple.V1.seqNr)
	require.False(t, tuple.V1.GetMembers().FindByUserAccountId(userAccountId).IsPresent())
	require.Equal(t, groupChat.id, tuple.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple.V2.GetSeqNr())
}

func TestGroupChat_Rename(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId, adminId)
	name := models.NewGroupChatName("test2").MustGet()

	// When
	result := groupChat.Rename(name, adminId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()
	require.Equal(t, groupChat.id, tuple.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple.V1.seqNr)
	require.Equal(t, "test2", tuple.V1.GetName().String())
	require.Equal(t, groupChat.id, tuple.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple.V2.GetSeqNr())
}

func TestGroupChat_Delete(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	groupChat, _ := NewGroupChat(groupChatName, adminId, adminId)

	// When
	result := groupChat.Delete(adminId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()
	require.Equal(t, groupChat.id, tuple.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple.V1.seqNr)
	require.True(t, tuple.V1.IsDeleted())
	require.Equal(t, groupChat.id, tuple.V2.GetAggregateId())
	require.Equal(t, groupChat.seqNr+1, tuple.V2.GetSeqNr())
}

func TestGroupChat_PostMessage(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	userAccountId := models.NewUserAccountId()
	groupChat, _ := NewGroupChat(groupChatName, adminId, adminId)
	groupChat = groupChat.AddMember(models.NewMemberId(), userAccountId, models.MemberRole, adminId).MustGet().V1
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", userAccountId)

	// When
	result := groupChat.PostMessage(message, userAccountId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()
	require.Equal(t, groupChat.id, tuple.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple.V1.seqNr)
	require.True(t, tuple.V1.GetMessages().Get(messageId).IsPresent())
	require.Equal(t, message, tuple.V1.GetMessages().Get(messageId).MustGet())
}

func TestGroupChat_DeleteMessage(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	userAccountId := models.NewUserAccountId()
	groupChat, _ := NewGroupChat(groupChatName, adminId, adminId)
	groupChat = groupChat.AddMember(models.NewMemberId(), userAccountId, models.MemberRole, adminId).MustGet().V1
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", userAccountId)
	result1 := groupChat.PostMessage(message, userAccountId)
	tuple1 := result1.MustGet()

	// When
	result2 := tuple1.V1.DeleteMessage(messageId, userAccountId)

	// Then
	require.True(t, result2.IsOk())
	tuple2 := result2.MustGet()
	require.Equal(t, groupChat.id, tuple2.V1.id)
	require.Equal(t, groupChat.seqNr+2, tuple2.V1.seqNr)
	require.True(t, tuple2.V1.GetMessages().Get(messageId).IsAbsent())
}
