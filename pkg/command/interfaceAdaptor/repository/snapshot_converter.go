package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

func SnapshotConverter(m map[string]interface{}) (esa.Aggregate, error) {
	groupChatId := models2.ConvertGroupChatIdFromJSON(m["Id"].(map[string]interface{}))
	name, err := models2.ConvertGroupChatNameFromJSON(m["Name"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	members := models2.ConvertMembersFromJSON(m["Members"].(map[string]interface{}))
	messages, err := models2.ConvertMessagesFromJSON(m["Messages"].(map[string]interface{})).Get()
	if err != nil {
		return nil, err
	}
	seqNr := uint64(m["SeqNr"].(float64))
	version := uint64(m["Version"].(float64))
	deleted := m["Deleted"].(bool)
	result := domain.NewGroupChatFrom(groupChatId, name, members, messages, seqNr, version, deleted)
	return result, nil
}
