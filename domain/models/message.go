package models

type Message struct {
	id       MessageId
	text     string
	senderId uint64
}

func NewMessage(id MessageId, text string, senderId uint64) *Message {
	return &Message{id: id, text: text, senderId: senderId}
}

func (m *Message) GetId() MessageId {
	return m.id
}

func (m *Message) GetText() string {
	return m.text
}

func (m *Message) GetSenderId() uint64 {
	return m.senderId
}
