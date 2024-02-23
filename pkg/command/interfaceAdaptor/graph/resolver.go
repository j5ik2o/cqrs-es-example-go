package commandgraph

import "cqrs-es-example-go/pkg/command/useCase"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	groupChatCommandProcessor useCase.GroupChatCommandProcessor
}

func NewResolver(groupChatCommandProcessor useCase.GroupChatCommandProcessor) *Resolver {
	return &Resolver{groupChatCommandProcessor}
}
