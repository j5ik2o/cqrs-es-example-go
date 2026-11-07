package repository

import (
	"cqrs-es-example-go/pkg/command/domain/events"
	"encoding/json"

	esa "github.com/j5ik2o/event-store-adapter-go/v5"
)

type EventSerializer struct{}

func NewEventSerializer() *EventSerializer {
	return &EventSerializer{}
}

func (s *EventSerializer) Serialize(event events.GroupChatEvent) ([]byte, error) {
	result, err := json.Marshal(event.ToJSON())
	if err != nil {
		return nil, &esa.SerializationError{Cause: err}
	}
	return result, nil
}

func (s *EventSerializer) Deserialize(data []byte) (events.GroupChatEvent, error) {
	eventMap, err := decodeJSONObject(data)
	if err != nil {
		return nil, &esa.SerializationError{Cause: err}
	}
	event, err := EventConverter(eventMap)
	if err != nil {
		return nil, &esa.SerializationError{Cause: err}
	}
	return event, nil
}
