package models

import (
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

// MessageId is a value object that represents a message id.
type MessageId struct {
	value string
}

// NewMessageId is the constructor for MessageId with generating id.
func NewMessageId() MessageId {
	id := ulid.Make()
	return MessageId{value: id.String()}
}

// NewMessageIdFromString is the constructor for MessageId.
func NewMessageIdFromString(value string) mo.Result[MessageId] {
	return mo.Ok(MessageId{value})
}

// ConvertMessageIdFromJSON is a constructor for MessageId.
func ConvertMessageIdFromJSON(value map[string]interface{}) MessageId {
	return NewMessageIdFromString(value["value"].(string)).MustGet()
}

// ToJSON converts to JSON.
//
// However, this method is out of layer.
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
