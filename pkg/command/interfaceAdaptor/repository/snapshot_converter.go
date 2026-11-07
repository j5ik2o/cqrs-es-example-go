package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
)

// SnapshotConverter restores aggregate state from its JSON payload. Unknown
// fields are ignored; a missing required field is reported as an error.
func SnapshotConverter(m map[string]interface{}) (domain.GroupChat, error) {
	idValue, err := jsonValueString(m, "id")
	if err != nil {
		return domain.GroupChat{}, err
	}
	groupChatId, err := models.NewGroupChatIdFromString(idValue).Get()
	if err != nil {
		return domain.GroupChat{}, err
	}
	name, err := convertGroupChatName(m)
	if err != nil {
		return domain.GroupChat{}, err
	}
	membersObject, err := jsonObject(m, "members")
	if err != nil {
		return domain.GroupChat{}, err
	}
	members, err := convertMembers(membersObject)
	if err != nil {
		return domain.GroupChat{}, err
	}
	messagesObject, err := jsonObject(m, "messages")
	if err != nil {
		return domain.GroupChat{}, err
	}
	messages, err := convertMessages(messagesObject)
	if err != nil {
		return domain.GroupChat{}, err
	}
	seqNrValue, err := jsonInteger(m, "seq_nr")
	if err != nil {
		return domain.GroupChat{}, err
	}
	if seqNrValue < 0 {
		return domain.GroupChat{}, fmt.Errorf("seq_nr must not be negative")
	}
	deleted, err := jsonBoolean(m, "deleted")
	if err != nil {
		return domain.GroupChat{}, err
	}
	return domain.NewGroupChatFrom(groupChatId, name, members, messages, uint64(seqNrValue), deleted), nil
}
