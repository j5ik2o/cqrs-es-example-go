package models

type Message struct {
	id       *MessageId
	text     string
	senderId *UserAccountId
}

func NewMessage(id *MessageId, text string, senderId *UserAccountId) *Message {
	return &Message{id, text, senderId}
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
