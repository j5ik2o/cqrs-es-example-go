package domain

import (
	"cqrs-es-example-go/domain/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGroupChat_AddMember(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test")
	members := models.NewMembers(adminId)
	groupChat := NewGroupChat(groupChatName).WithMembers(members)
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
	groupChatName := models.NewGroupChatName("test")
	members := models.NewMembers(adminId)
	groupChat := NewGroupChat(groupChatName).WithMembers(members)
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
	groupChatName := models.NewGroupChatName("test")
	members := models.NewMembers(adminId)
	groupChat := NewGroupChat(groupChatName).WithMembers(members)

	// When
	result := groupChat.Rename(models.NewGroupChatName("test2"), adminId)

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
	groupChatName := models.NewGroupChatName("test")
	members := models.NewMembers(adminId)
	groupChat := NewGroupChat(groupChatName).WithMembers(members)

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
	groupChatName := models.NewGroupChatName("test")
	userAccountId := models.NewUserAccountId()
	members := models.NewMembers(adminId).AddMember(userAccountId)
	groupChat := NewGroupChat(groupChatName).WithMembers(members)
	messageId := models.NewMessageId()
	message := models.NewMessage(messageId, "test", userAccountId)

	// When
	result := groupChat.PostMessage(message, adminId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()

	require.Equal(t, groupChat.id, tuple.V1.id)
	require.Equal(t, groupChat.seqNr+1, tuple.V1.seqNr)
	// require.True(t, tuple.V1.GetMessages().Get(messageId).IsPresent())
}
