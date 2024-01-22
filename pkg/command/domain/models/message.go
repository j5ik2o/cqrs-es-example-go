package models

import "github.com/samber/mo"

type Message struct {
	id       *MessageId
	text     string
	senderId *UserAccountId
}

func NewMessage(id *MessageId, text string, senderId *UserAccountId) mo.Result[*Message] {
	return mo.Ok(&Message{id, text, senderId})
}

func ConvertMessageFromJSON(value map[string]interface{}) mo.Result[*Message] {
	return NewMessage(
		ConvertMessageIdFromJSON(value["id"].(map[string]interface{})),
		value["text"].(string),
		ConvertUserAccountIdFromJSON(value["sender_id"].(map[string]interface{})).MustGet(),
	)
}

func (m *Message) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":        m.id.ToJSON(),
		"text":      m.text,
		"sender_id": m.senderId.ToJSON(),
	}
}

func (m *Message) GetId() *MessageId {
	return m.id
}

func (m *Message) GetText() string {
	return m.text
}

func (m *Message) GetSenderId() *UserAccountId {
	return m.senderId
}

func (m *Message) Equals(other *Message) bool {
	return m.id.Equals(other.id)
}
