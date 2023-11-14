package models

import "github.com/samber/mo"

type Messages struct {
	values map[*MessageId]*Message
}

func NewMessages() *Messages {
	return &Messages{values: map[*MessageId]*Message{}}
}

func NewMessagesFromMap(values map[*MessageId]*Message) *Messages {
	return &Messages{values: values}
}

func ConvertMessagesFromJSON(value map[string]interface{}) mo.Result[*Messages] {
	values := value["Values"].([]interface{})
	m := NewMessages()
	for _, v := range values {
		result := ConvertMessageFromJSON(v.(map[string]interface{}))
		if result.IsError() {
			return mo.Err[*Messages](result.Error())
		}
		m = m.Add(result.MustGet()).MustGet()
	}
	return mo.Ok(m)
}

func (m *Messages) Contains(id *MessageId) bool {
	_, ok := m.values[id]
	return ok
}

func (m *Messages) Add(message *Message) mo.Option[*Messages] {
	if m.Contains(message.GetId()) {
		return mo.None[*Messages]()
	}
	newMap := m.ToMap()
	newMap[message.GetId()] = message
	return mo.Some(NewMessagesFromMap(newMap))
}

func (m *Messages) Remove(id *MessageId) mo.Option[*Messages] {
	if !m.Contains(id) {
		return mo.None[*Messages]()
	}
	newMap := m.ToMap()
	delete(newMap, id)
	return mo.Some(NewMessagesFromMap(newMap))
}

func (m *Messages) Get(id *MessageId) mo.Option[*Message] {
	if !m.Contains(id) {
		return mo.None[*Message]()
	}
	return mo.Some[*Message](m.values[id])
}

func (m *Messages) ToJSON() map[string]interface{} {
	values := make([]map[string]interface{}, len(m.values))
	i := 0
	for _, v := range m.values {
		values[i] = v.ToJSON()
		i++
	}
	return map[string]interface{}{
		"Values": values,
	}
}

func (m *Messages) ToMap() map[*MessageId]*Message {
	result := map[*MessageId]*Message{}
	for k, v := range m.values {
		result[k] = v
	}
	return result
}

func (m *Messages) ToSlice() []*Message {
	result := make([]*Message, len(m.values))
	i := 0
	for _, v := range m.values {
		result[i] = v
		i++
	}
	return result
}
