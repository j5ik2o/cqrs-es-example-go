package rmu

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// JournalEvent is one event read from the journal table. It carries the same
// envelope metadata as a head event.
type JournalEvent struct {
	AggregateID string
	SeqNr       int64
	OccurredAt  time.Time
	Manifest    string
	Payload     []byte
}

// JournalReader fills a sequence number gap detected in the head stream. The
// journal is read with strong consistency (DY-11).
type JournalReader interface {
	// ReadEvents returns the events with fromSeqNr <= seq_nr <= toSeqNr in
	// ascending order.
	ReadEvents(ctx context.Context, aggregateID string, fromSeqNr, toSeqNr int64) ([]JournalEvent, error)
}

type DynamoDBJournalReader struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBJournalReader(client *dynamodb.Client, tableName string) *DynamoDBJournalReader {
	return &DynamoDBJournalReader{client: client, tableName: tableName}
}

func (r *DynamoDBJournalReader) ReadEvents(ctx context.Context, aggregateID string, fromSeqNr, toSeqNr int64) ([]JournalEvent, error) {
	if fromSeqNr > toSeqNr {
		return nil, nil
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		ConsistentRead:         aws.Bool(true),
		ScanIndexForward:       aws.Bool(true),
		KeyConditionExpression: aws.String("aid = :aid AND seq_nr BETWEEN :from_seq_nr AND :to_seq_nr"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":aid":         &types.AttributeValueMemberS{Value: aggregateID},
			":from_seq_nr": &types.AttributeValueMemberN{Value: strconv.FormatInt(fromSeqNr, 10)},
			":to_seq_nr":   &types.AttributeValueMemberN{Value: strconv.FormatInt(toSeqNr, 10)},
		},
	}
	var events []JournalEvent
	for {
		out, err := r.client.Query(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, item := range out.Items {
			event, err := journalEventFromItem(item)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
		if len(out.LastEvaluatedKey) == 0 {
			return events, nil
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}
}

func journalEventFromItem(item map[string]types.AttributeValue) (JournalEvent, error) {
	aid, ok := item["aid"].(*types.AttributeValueMemberS)
	if !ok {
		return JournalEvent{}, fmt.Errorf("journal item has no aid string")
	}
	seqNr, ok := item["seq_nr"].(*types.AttributeValueMemberN)
	if !ok {
		return JournalEvent{}, fmt.Errorf("journal item %s has no seq_nr number", aid.Value)
	}
	parsedSeqNr, err := strconv.ParseInt(seqNr.Value, 10, 64)
	if err != nil {
		return JournalEvent{}, err
	}
	occurredAt, ok := item["occurred_at"].(*types.AttributeValueMemberN)
	if !ok {
		return JournalEvent{}, fmt.Errorf("journal item %s has no occurred_at number", aid.Value)
	}
	occurredAtNanos, err := strconv.ParseInt(occurredAt.Value, 10, 64)
	if err != nil {
		return JournalEvent{}, err
	}
	manifest, ok := item["manifest"].(*types.AttributeValueMemberS)
	if !ok {
		return JournalEvent{}, fmt.Errorf("journal item %s has no manifest string", aid.Value)
	}
	payload, err := payloadBytes(item["payload"])
	if err != nil {
		return JournalEvent{}, err
	}
	return JournalEvent{
		AggregateID: aid.Value,
		SeqNr:       parsedSeqNr,
		OccurredAt:  time.Unix(0, occurredAtNanos),
		Manifest:    manifest.Value,
		Payload:     payload,
	}, nil
}

func payloadBytes(value types.AttributeValue) ([]byte, error) {
	switch payload := value.(type) {
	case *types.AttributeValueMemberB:
		return payload.Value, nil
	case *types.AttributeValueMemberS:
		return []byte(payload.Value), nil
	default:
		return nil, fmt.Errorf("journal item has no payload")
	}
}
