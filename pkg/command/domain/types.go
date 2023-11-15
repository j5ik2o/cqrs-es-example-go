package domain

import (
	"cqrs-es-example-go/pkg/command/domain/events"
	gt "github.com/barweiss/go-tuple"
)

type GroupChatWithEventPair = gt.Pair[*GroupChat, events.GroupChatEvent]
