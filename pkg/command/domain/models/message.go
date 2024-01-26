package models

import "github.com/samber/mo"

// Message is a value object that represents a message.
type Message struct {
	id       *MessageId
	text     string
	senderId UserAccountId
}

// NewMessage is the constructor for Message.
func NewMessage(id *MessageId, text string, senderId UserAccountId) mo.Result[*Message] {
	return mo.Ok(&Message{id, text, senderId})
}

// ConvertMessageFromJSON is a constructor for Message.
func ConvertMessageFromJSON(value map[string]interface{}) mo.Result[*Message] {
	return NewMessage(
		ConvertMessageIdFromJSON(value["id"].(map[string]interface{})),
		value["text"].(string),
		ConvertUserAccountIdFromJSON(value["sender_id"].(map[string]interface{})).MustGet(),
	)
}

// ToJSON converts to JSON.
func (m *Message) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":        m.id.ToJSON(),
		"text":      m.text,
		"sender_id": m.senderId.ToJSON(),
	}
}

// GetId returns id.
func (m *Message) GetId() *MessageId {
	return m.id
}

// GetText returns text.
func (m *Message) GetText() string {
	return m.text
}

// GetSenderId returns sender id.
func (m *Message) GetSenderId() *UserAccountId {
	return &m.senderId
}

// Equals returns whether the message is equal to the other.
func (m *Message) Equals(other *Message) bool {
	return m.id.Equals(other.id) && m.text == other.text && m.senderId.Equals(&other.senderId)
}
