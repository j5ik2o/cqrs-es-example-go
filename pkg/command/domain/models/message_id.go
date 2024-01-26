package models

import (
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

type MessageId struct {
	value string
}

func NewMessageId() MessageId {
	id := ulid.Make()
	return MessageId{value: id.String()}
}

func NewMessageIdFromString(value string) mo.Result[MessageId] {
	return mo.Ok(MessageId{value})
}

func ConvertMessageIdFromJSON(value map[string]interface{}) MessageId {
	return NewMessageIdFromString(value["value"].(string)).MustGet()
}

func (m *MessageId) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func (m *MessageId) GetValue() string {
	return m.value
}

func (m *MessageId) Equals(other *MessageId) bool {
	return m.value == other.value
}

func (m *MessageId) String() string {
	return m.value
}
