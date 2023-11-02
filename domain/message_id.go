package domain

import "github.com/oklog/ulid/v2"

type MessageId struct {
	value string
}

func NewMessageId() *MessageId {
	id := ulid.Make()
	return &MessageId{value: id.String()}
}

func NewMessageIdFromString(value string) *MessageId {
	return &MessageId{value}
}

func (m *MessageId) GetValue() string {
	return m.value
}

func (m *MessageId) String() string {
	return m.value
}
