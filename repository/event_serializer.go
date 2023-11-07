package repository

import (
	"cqrs-es-example-go/domain/events"
	"encoding/json"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

type EventSerializer struct{}

func (s *EventSerializer) Serialize(event esa.Event) ([]byte, error) {
	result, err := json.Marshal(event.(events.GroupChatEvent).ToJSON())
	if err != nil {
		return nil, esa.NewSerializationError("Failed to serialize the event", err)
	}
	return result, nil
}

func (s *EventSerializer) Deserialize(data []byte, eventMap *map[string]interface{}) error {
	err := json.Unmarshal(data, eventMap)
	if err != nil {
		return esa.NewDeserializationError("Failed to deserialize the event", err)
	}
	return nil
}
