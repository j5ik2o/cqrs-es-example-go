package domain

import (
	"cqrs-es-example-go/domain/events"
	gt "github.com/barweiss/go-tuple"
)

type GroupChatWithEventPair = gt.Pair[*GroupChat, events.GroupChatEvent]
