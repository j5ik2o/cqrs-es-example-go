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
		ConvertMessageIdFromJSON(value["Id"].(map[string]interface{})),
		value["Text"].(string),
		ConvertUserAccountIdFromJSON(value["SenderId"].(map[string]interface{})).MustGet(),
	)
}

func (m *Message) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Id":       m.id.ToJSON(),
		"Text":     m.text,
		"SenderId": m.senderId.ToJSON(),
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
