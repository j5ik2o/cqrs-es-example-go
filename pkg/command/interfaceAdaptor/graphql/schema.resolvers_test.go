package commandgraphql

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	commandgraphql "cqrs-es-example-go/pkg/command/interfaceAdaptor/graphql/model"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/repository"
	"cqrs-es-example-go/pkg/command/processor"
	"cqrs-es-example-go/test"
	esa "github.com/j5ik2o/event-store-adapter-go/pkg"
	"github.com/j5ik2o/event-store-adapter-go/pkg/common"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"testing"
)

var (
	journalTableName     = "journal"
	journalAidIndexName  = "journal-aid-index"
	snapshotTableName    = "snapshot"
	snapshotAidIndexName = "snapshot-aid-index"
)

func Test_CreateGroupChat(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	result, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(result.GroupChatID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	require.True(t, name.Equals(actualGroupChat.GetName()))
	require.True(t, adminId.Equals(actualGroupChat.GetMembers().GetAdministrator().GetUserAccountId()))
}

func Test_RenameGroupChat(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	result, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatID := result.GroupChatID
	newName := models.NewGroupChatName("newName").MustGet()

	_, err = resolver.MutationRoot().RenameGroupChat(ctx, commandgraphql.RenameGroupChatInput{
		GroupChatID: groupChatID,
		Name:        newName.String(),
		ExecutorID:  adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(result.GroupChatID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	require.True(t, newName.Equals(actualGroupChat.GetName()))
}

func Test_AddMember(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	result, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatID := result.GroupChatID
	userAccountId := models.NewUserAccountId()
	result, err = resolver.MutationRoot().AddMember(ctx, commandgraphql.AddMemberInput{
		GroupChatID:   groupChatID,
		UserAccountID: userAccountId.AsString(),
		Role:          commandgraphql.RoleMember,
		ExecutorID:    adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(result.GroupChatID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	actualMember := actualGroupChat.GetMembers().FindByUserAccountId(&userAccountId).MustGet()
	require.True(t, actualMember.GetUserAccountId().Equals(&userAccountId))
	require.Equal(t, "member", actualMember.GetRole().String())
}

func Test_RemoveMember(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	result, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatID := result.GroupChatID
	userAccountId := models.NewUserAccountId()
	result, err = resolver.MutationRoot().AddMember(ctx, commandgraphql.AddMemberInput{
		GroupChatID:   groupChatID,
		UserAccountID: userAccountId.AsString(),
		Role:          commandgraphql.RoleMember,
		ExecutorID:    adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	result, err = resolver.MutationRoot().RemoveMember(ctx, commandgraphql.RemoveMemberInput{
		GroupChatID:   groupChatID,
		UserAccountID: userAccountId.AsString(),
		ExecutorID:    adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(result.GroupChatID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	require.False(t, actualGroupChat.GetMembers().ContainsByUserAccountId(&userAccountId))
}

func Test_PostMessage(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	groupResult, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	groupChatID := groupResult.GroupChatID
	userAccountId := models.NewUserAccountId()
	groupResult, err = resolver.MutationRoot().AddMember(ctx, commandgraphql.AddMemberInput{
		GroupChatID:   groupChatID,
		UserAccountID: userAccountId.AsString(),
		Role:          commandgraphql.RoleMember,
		ExecutorID:    adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	content := "test message"
	messageResult, err := resolver.MutationRoot().PostMessage(ctx, commandgraphql.PostMessageInput{
		GroupChatID: groupChatID,
		Content:     content,
		ExecutorID:  userAccountId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, messageResult)
	require.NotNil(t, messageResult.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(messageResult.GroupChatID).MustGet()
	messageId := models.NewMessageIdFromString(messageResult.MessageID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	message := actualGroupChat.GetMessages().Get(&messageId).MustGet()
	require.Equal(t, message.GetText(), content)
}

func Test_EditMessage(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	groupResult, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	groupChatID := groupResult.GroupChatID
	userAccountId := models.NewUserAccountId()
	groupResult, err = resolver.MutationRoot().AddMember(ctx, commandgraphql.AddMemberInput{
		GroupChatID:   groupChatID,
		UserAccountID: userAccountId.AsString(),
		Role:          commandgraphql.RoleMember,
		ExecutorID:    adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	content := "test message"
	messageResult, err := resolver.MutationRoot().PostMessage(ctx, commandgraphql.PostMessageInput{
		GroupChatID: groupChatID,
		Content:     content,
		ExecutorID:  userAccountId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, messageResult)
	require.NotNil(t, messageResult.GroupChatID)

	newContent := "new message"

	groupResult, err = resolver.MutationRoot().EditMessage(ctx, commandgraphql.EditMessageInput{
		GroupChatID: groupChatID,
		MessageID:   messageResult.MessageID,
		Content:     newContent,
		ExecutorID:  userAccountId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, messageResult)
	require.NotNil(t, messageResult.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(messageResult.GroupChatID).MustGet()
	messageId := models.NewMessageIdFromString(messageResult.MessageID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	message := actualGroupChat.GetMessages().Get(&messageId).MustGet()
	require.Equal(t, message.GetText(), newContent)
}

func Test_DeleteMessage(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)

	groupChatRepository, resolver, err := createResolver(t, ctx, container)
	require.NoError(t, err)

	adminId := models.NewUserAccountId()
	name := models.NewGroupChatName("test").MustGet()

	groupResult, err := resolver.MutationRoot().CreateGroupChat(ctx, commandgraphql.CreateGroupChatInput{
		Name:       name.String(),
		ExecutorID: adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	groupChatID := groupResult.GroupChatID
	userAccountId := models.NewUserAccountId()
	groupResult, err = resolver.MutationRoot().AddMember(ctx, commandgraphql.AddMemberInput{
		GroupChatID:   groupChatID,
		UserAccountID: userAccountId.AsString(),
		Role:          commandgraphql.RoleMember,
		ExecutorID:    adminId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	message := "test message"
	messageResult, err := resolver.MutationRoot().PostMessage(ctx, commandgraphql.PostMessageInput{
		GroupChatID: groupChatID,
		Content:     message,
		ExecutorID:  userAccountId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, messageResult)
	require.NotNil(t, messageResult.GroupChatID)

	messageId := models.NewMessageIdFromString(messageResult.MessageID).MustGet()

	groupResult, err = resolver.MutationRoot().DeleteMessage(ctx, commandgraphql.DeleteMessageInput{
		GroupChatID: groupChatID,
		MessageID:   messageId.String(),
		ExecutorID:  userAccountId.AsString(),
	})
	require.NoError(t, err)
	require.NotNil(t, groupResult)
	require.NotNil(t, groupResult.GroupChatID)

	groupChatId := models.NewGroupChatIdFromString(messageResult.GroupChatID).MustGet()
	actualGroupChat, b := groupChatRepository.FindById(&groupChatId).MustGet().Get()
	require.True(t, b)
	require.NotNil(t, actualGroupChat)
	require.True(t, groupChatId.Equals(actualGroupChat.GetGroupChatId()))
	require.False(t, actualGroupChat.GetMessages().Contains(&messageId))
}

func createResolver(t *testing.T, ctx context.Context, container *localstack.LocalStackContainer) (*repository.GroupChatRepositoryImpl, *Resolver, error) {
	dynamodbClient, err := common.CreateDynamoDBClient(t, ctx, container)
	if err != nil {
		return nil, nil, err
	}

	err = common.CreateJournalTable(t, ctx, dynamodbClient, journalTableName, journalAidIndexName)
	if err != nil {
		return nil, nil, err
	}

	err = common.CreateSnapshotTable(t, ctx, dynamodbClient, snapshotTableName, snapshotAidIndexName)
	if err != nil {
		return nil, nil, err
	}

	eventStore, err := esa.NewEventStoreOnDynamoDB(
		dynamodbClient,
		journalTableName, snapshotTableName, journalAidIndexName, snapshotAidIndexName,
		32,
		repository.EventConverter,
		repository.SnapshotConverter,
		esa.WithEventSerializer(repository.NewEventSerializer()),
		esa.WithSnapshotSerializer(repository.NewSnapshotSerializer()))
	if err != nil {
		return nil, nil, err
	}

	groupChatRepository, err := repository.NewGroupChatRepository(eventStore, repository.WithRetention(2))
	if err != nil {
		return nil, nil, err
	}

	groupChatCommandProcessor := processor.NewGroupChatCommandProcessor(&groupChatRepository)

	resolver := NewResolver(groupChatCommandProcessor)
	return &groupChatRepository, resolver, err
}
