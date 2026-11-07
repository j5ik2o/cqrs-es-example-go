package repository

import (
	"cqrs-es-example-go/pkg/command/domain"
	"encoding/json"

	esa "github.com/j5ik2o/event-store-adapter-go/v5"
)

type SnapshotSerializer struct{}

func NewSnapshotSerializer() *SnapshotSerializer {
	return &SnapshotSerializer{}
}

func (s *SnapshotSerializer) Serialize(aggregate domain.GroupChat) ([]byte, error) {
	result, err := json.Marshal(aggregate.ToJSON())
	if err != nil {
		return nil, &esa.SerializationError{Cause: err}
	}
	return result, nil
}

func (s *SnapshotSerializer) Deserialize(data []byte) (domain.GroupChat, error) {
	aggregateMap, err := decodeJSONObject(data)
	if err != nil {
		return domain.GroupChat{}, &esa.SerializationError{Cause: err}
	}
	aggregate, err := SnapshotConverter(aggregateMap)
	if err != nil {
		return domain.GroupChat{}, &esa.SerializationError{Cause: err}
	}
	return aggregate, nil
}
