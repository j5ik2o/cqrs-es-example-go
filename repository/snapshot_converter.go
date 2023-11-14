package repository

import (
	"cqrs-es-example-go/domain"
	"cqrs-es-example-go/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

func SnapshotConverter(m map[string]interface{}) (esa.Aggregate, error) {
	groupChatId := models.ConvertGroupChatIdFromJSON(m["Id"].(map[string]interface{}))
	name, err := models.ConvertGroupChatNameFromJSON(m["Name"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	members := models.ConvertMembersFromJSON(m["Members"].(map[string]interface{}))
	messages, err := models.ConvertMessagesFromJSON(m["Messages"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	seqNr := uint64(m["SeqNr"].(float64))
	version := uint64(m["Version"].(float64))
	deleted := m["Deleted"].(bool)
	result := domain.NewGroupChatFrom(groupChatId, name, members, messages, seqNr, version, deleted)
	return result, nil
}
