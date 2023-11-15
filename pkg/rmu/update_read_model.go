package rmu

import (
	"context"
	"fmt"
)
import "github.com/aws/aws-lambda-go/events"

func UpdateReadModel(ctx context.Context, event events.DynamoDBEvent) {
	for _, record := range event.Records {
		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)

		// Print new values for attributes of type String
		for name, value := range record.Change.NewImage {
			fmt.Printf("Attribute name: %s, value: %s\n", name, value.String())
		}
	}
}
