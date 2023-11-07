package repository

import (
	"cqrs-es-example-go/domain"
	"encoding/json"
	esa "github.com/j5ik2o/event-store-adapter-go"
)

type SnapshotSerializer struct{}

func (s *SnapshotSerializer) Serialize(aggregate esa.Aggregate) ([]byte, error) {
	result, err := json.Marshal(aggregate.(*domain.GroupChat).ToJSON())
	if err != nil {
		return nil, esa.NewSerializationError("Failed to serialize the snapshot", err)
	}
	return result, nil
}

func (s *SnapshotSerializer) Deserialize(data []byte, aggregateMap *map[string]interface{}) error {
	err := json.Unmarshal(data, aggregateMap)
	if err != nil {
		return esa.NewDeserializationError("Failed to deserialize the snapshot", err)
	}
	return nil
}
