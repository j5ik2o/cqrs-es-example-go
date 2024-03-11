package models

import (
	"cqrs-es-example-go/pkg/command/domain/errors"
	"github.com/samber/mo"
)

// Messages is a value object that represents a list of messages.
type Messages struct {
	values map[string]Message
}

// NewMessages is the constructor for Messages.
func NewMessages() Messages {
	return Messages{values: map[string]Message{}}
}

// NewMessagesFromMap is the constructor for Messages.
func NewMessagesFromMap(values map[string]Message) Messages {
	m := make(map[string]Message, len(values))
	for k, v := range values {
		m[k] = v
	}
	return Messages{values: m}
}

// ConvertMessagesFromJSON is a constructor for Messages.
func ConvertMessagesFromJSON(value map[string]interface{}) mo.Result[Messages] {
	values := value["values"].([]interface{})
	m := NewMessages()
	for _, v := range values {
		result := ConvertMessageFromJSON(v.(map[string]interface{}))
		if result.IsError() {
			return mo.Err[Messages](result.Error())
		}
		m = m.Add(result.MustGet()).MustGet()
	}
	return mo.Ok(m)
}

func (m *Messages) Contains(id *MessageId) bool {
	_, ok := m.values[id.GetValue()]
	return ok
}

func (m *Messages) Add(message Message) mo.Result[Messages] {
	if m.Contains(message.GetId()) {
		return mo.Err[Messages](errors.NewAlreadyExistsMessageError("message already exists: " + message.GetId().GetValue()))
	}
	newMap := m.ToMap()
	newMap[message.GetId().GetValue()] = message
	return mo.Ok(NewMessagesFromMap(newMap))
}

func (m *Messages) Edit(message Message) mo.Result[Messages] {
	if !m.Contains(message.GetId()) {
		return mo.Err[Messages](errors.NewMessageNotFoundError("message not found: " + message.GetId().GetValue()))
	}
	if m.values[message.GetId().GetValue()].senderId != message.senderId {
		return mo.Err[Messages](errors.NewNotSenderError("not authorized: " + message.GetId().GetValue()))
	}
	newMap := m.ToMap()
	newMap[message.GetId().GetValue()] = message
	return mo.Ok(NewMessagesFromMap(newMap))
}

func (m *Messages) Remove(id *MessageId, senderId UserAccountId) mo.Result[Messages] {
	if !m.Contains(id) {
		return mo.Err[Messages](errors.NewMessageNotFoundError("message not found: " + id.GetValue()))
	}
	if m.values[id.GetValue()].senderId != senderId {
		return mo.Err[Messages](errors.NewNotSenderError("not authorized: " + id.GetValue()))
	}
	newMap := m.ToMap()
	delete(newMap, id.GetValue())
	return mo.Ok(NewMessagesFromMap(newMap))
}

func (m *Messages) Get(id *MessageId) mo.Option[Message] {
	if !m.Contains(id) {
		return mo.None[Message]()
	}
	return mo.Some[Message](m.values[id.GetValue()])
}

// ToJSON converts to JSON.
//
// However, this method is out of layer.
func (m *Messages) ToJSON() map[string]interface{} {
	values := make([]map[string]interface{}, len(m.values))
	i := 0
	for _, v := range m.values {
		values[i] = v.ToJSON()
		i++
	}
	return map[string]interface{}{
		"values": values,
	}
}

func (m *Messages) ToMap() map[string]Message {
	result := map[string]Message{}
	for k, v := range m.values {
		result[k] = v
	}
	return result
}

func (m *Messages) ToSlice() []Message {
	result := make([]Message, len(m.values))
	i := 0
	for _, v := range m.values {
		result[i] = v
		i++
	}
	return result
}
