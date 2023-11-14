package models

type Message struct {
	id       *MessageId
	text     string
	senderId *UserAccountId
}

func NewMessage(id *MessageId, text string, senderId *UserAccountId) *Message {
	return &Message{id, text, senderId}
}

func ConvertMessageFromJSON(value map[string]interface{}) *Message {
	json, err := ConvertUserAccountIdFromJSON(value["SenderId"].(map[string]interface{}))
	if err != nil {
		panic(err)
	}
	return NewMessage(
		ConvertMessageIdFromJSON(value["Id"].(map[string]interface{})),
		value["Text"].(string),
		json,
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
