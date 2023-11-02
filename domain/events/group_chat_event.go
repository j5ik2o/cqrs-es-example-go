package events

import (
	esa "github.com/j5ik2o/event-store-adapter-go"
)

type GroupChatEvent interface {
	esa.Event
}
