package commandgraphql

import "cqrs-es-example-go/pkg/command/processor"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	groupChatCommandProcessor processor.GroupChatCommandProcessor
}

func NewResolver(groupChatCommandProcessor processor.GroupChatCommandProcessor) *Resolver {
	return &Resolver{groupChatCommandProcessor}
}
