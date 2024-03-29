package commandgraphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	commandgraphql "cqrs-es-example-go/pkg/command/interfaceAdaptor/graphql/model"
	"cqrs-es-example-go/pkg/command/interfaceAdaptor/validators"
	"fmt"
)

// CreateGroupChat is the resolver for the createGroupChat field.
func (r *mutationRootResolver) CreateGroupChat(ctx context.Context, input commandgraphql.CreateGroupChatInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatName, err := validators.ValidateGroupChatName(input.Name).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.CreateGroupChat(groupChatName, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// DeleteGroupChat is the resolver for the deleteGroupChat field.
func (r *mutationRootResolver) DeleteGroupChat(ctx context.Context, input commandgraphql.DeleteGroupChatInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.DeleteGroupChat(&groupChatId, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// RenameGroupChat is the resolver for the renameGroupChat field.
func (r *mutationRootResolver) RenameGroupChat(ctx context.Context, input commandgraphql.RenameGroupChatInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	groupChatName, err := validators.ValidateGroupChatName(input.Name).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.RenameGroupChat(&groupChatId, groupChatName, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// AddMember is the resolver for the addMember field.
func (r *mutationRootResolver) AddMember(ctx context.Context, input commandgraphql.AddMemberInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	accountId, err := validators.ValidateUserAccountId(input.UserAccountID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	role, err := models.StringToRole(input.Role.String())
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.AddMember(&groupChatId, accountId, role, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// RemoveMember is the resolver for the removeMember field.
func (r *mutationRootResolver) RemoveMember(ctx context.Context, input commandgraphql.RemoveMemberInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	userAccountId, err := validators.ValidateUserAccountId(input.UserAccountID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.RemoveMember(&groupChatId, userAccountId, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// PostMessage is the resolver for the postMessage field.
func (r *mutationRootResolver) PostMessage(ctx context.Context, input commandgraphql.PostMessageInput) (*commandgraphql.MessageResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	messageId := models.NewMessageId()

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	message, err := validators.ValidateMessage(messageId, input.Content, executorId).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.PostMessage(&groupChatId, message, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.MessageResult{GroupChatID: event.GetAggregateId().AsString(), MessageID: messageId.String()}, nil
}

// EditMessage is the resolver for the editMessage field.
func (r *mutationRootResolver) EditMessage(ctx context.Context, input commandgraphql.EditMessageInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	messageId, err := validators.ValidateMessageId(input.MessageID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	message, err := validators.ValidateMessage(messageId, input.Content, executorId).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.EditMessage(&groupChatId, message, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// DeleteMessage is the resolver for the deleteMessage field.
func (r *mutationRootResolver) DeleteMessage(ctx context.Context, input commandgraphql.DeleteMessageInput) (*commandgraphql.GroupChatResult, error) {
	var errorList []error

	groupChatId, err := validators.ValidateGroupChatId(input.GroupChatID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	messageId, err := validators.ValidateMessageId(input.MessageID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	executorId, err := validators.ValidateUserAccountId(input.ExecutorID).Get()
	if err != nil {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		validationErrorHandling(ctx, errorList)
		return nil, nil
	}

	event, err := r.groupChatCommandProcessor.DeleteMessage(&groupChatId, messageId, executorId).Get()
	if err != nil {
		errorHandling(ctx, err)
		return nil, nil
	}

	return &commandgraphql.GroupChatResult{GroupChatID: event.GetAggregateId().AsString()}, nil
}

// HealthCheck is the resolver for the healthCheck field.
func (r *queryRootResolver) HealthCheck(ctx context.Context) (string, error) {
	panic(fmt.Errorf("not implemented: HealthCheck - healthCheck"))
}

// MutationRoot returns MutationRootResolver implementation.
func (r *Resolver) MutationRoot() MutationRootResolver { return &mutationRootResolver{r} }

// QueryRoot returns QueryRootResolver implementation.
func (r *Resolver) QueryRoot() QueryRootResolver { return &queryRootResolver{r} }

type mutationRootResolver struct{ *Resolver }
type queryRootResolver struct{ *Resolver }
