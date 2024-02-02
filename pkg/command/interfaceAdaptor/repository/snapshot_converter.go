package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	"cqrs-es-example-go/pkg/command/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

func SnapshotConverter(m map[string]interface{}) (esa.Aggregate, error) {
	groupChatId, err := models.ConvertGroupChatIdFromJSON(m["id"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	name, err := models.ConvertGroupChatNameFromJSON(m["name"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	members := models.ConvertMembersFromJSON(m["members"].(map[string]interface{}))
	messages, err := models.ConvertMessagesFromJSON(m["messages"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	seqNr := uint64(m["seq_nr"].(float64))
	version := uint64(m["version"].(float64))
	deleted := m["deleted"].(bool)
	result := domain.NewGroupChatFrom(groupChatId, name, members, messages, seqNr, version, deleted)
	return &result, nil
}
