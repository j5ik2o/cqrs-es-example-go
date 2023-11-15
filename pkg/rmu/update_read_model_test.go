package rmu

import (
	_ "embed"
	"encoding/json"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

//go:embed example-dynamodb-event.json
var eventData []byte

var (
	requestId  string
	deadlineMs string
	initOnce   sync.Once
)

func TestUpdateReadModel(t *testing.T) {
	var parsed dynamodbevents.DynamoDBEvent
	err := json.Unmarshal(eventData, &parsed)
	require.NoError(t, err)
	// UpdateReadModel(context.Background(), parsed)
	// Given
	// When
	// Then

}
