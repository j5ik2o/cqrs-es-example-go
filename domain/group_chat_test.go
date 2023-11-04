package domain

import (
	"cqrs-es-example-go/domain/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_AddMember(t *testing.T) {
	// Given
	adminId := models.NewUserAccountId()
	groupChatName := models.NewGroupChatName("test")
	members := models.NewMembers(adminId)
	gc := NewGroupChat(groupChatName, members)
	memberId := models.NewMemberId()
	userAccountId := models.NewUserAccountId()

	// When
	result := gc.AddMember(memberId, userAccountId, models.MemberRole, adminId)

	// Then
	require.True(t, result.IsOk())
	tuple, _ := result.Get()
	require.Equal(t, gc.id, tuple.V1.id)
	require.Equal(t, gc.seqNr+1, tuple.V1.seqNr)
	require.True(t, tuple.V1.GetMembers().FindByUserAccountId(userAccountId).IsPresent())
}
