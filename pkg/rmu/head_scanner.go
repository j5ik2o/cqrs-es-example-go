package rmu

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Head is one entry of the head table after the configuration item is removed.
type Head struct {
	AggregateID string
	TypeName    string
	SeqNr       int64
}

// HeadScanner walks the whole head table. DY-15 requires a strongly consistent
// read so a resync sees every aggregate that advanced while the reader was
// stopped.
type HeadScanner interface {
	ScanHeads(ctx context.Context, fn func(Head) error) error
}

// StoreIDProvider returns the store identifier shared by the three DynamoDB
// tables. Recovery tasks record it so a task cannot be replayed against another
// store's journal.
type StoreIDProvider interface {
	StoreID(ctx context.Context) (string, error)
}

// DynamoDBHeadScanner scans the head table. It keeps the configuration item out
// of the callback; the store identifier is read separately from __config__.
type DynamoDBHeadScanner struct {
	client    *dynamodb.Client
	tableName string
	pageLimit int32

	mu      sync.Mutex
	storeID string
}

type HeadScannerOption func(*DynamoDBHeadScanner)

// WithPageLimit caps one Scan page. Tests use it to force pagination; the
// default lets DynamoDB choose the page size.
func WithPageLimit(limit int32) HeadScannerOption {
	return func(s *DynamoDBHeadScanner) {
		s.pageLimit = limit
	}
}

func NewDynamoDBHeadScanner(client *dynamodb.Client, tableName string, opts ...HeadScannerOption) *DynamoDBHeadScanner {
	scanner := &DynamoDBHeadScanner{client: client, tableName: tableName}
	for _, opt := range opts {
		opt(scanner)
	}
	return scanner
}

func (s *DynamoDBHeadScanner) ScanHeads(ctx context.Context, fn func(Head) error) error {
	input := &dynamodb.ScanInput{
		TableName:            aws.String(s.tableName),
		ConsistentRead:       aws.Bool(true),
		ProjectionExpression: aws.String("aid, seq_nr, type_name"),
		FilterExpression:     aws.String("aid <> :config"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":config": &types.AttributeValueMemberS{Value: configAggregateID},
		},
	}
	if s.pageLimit > 0 {
		input.Limit = aws.Int32(s.pageLimit)
	}
	for {
		out, err := s.client.Scan(ctx, input)
		if err != nil {
			return err
		}
		for _, item := range out.Items {
			head, err := headFromItem(item)
			if err != nil {
				return err
			}
			if err := fn(head); err != nil {
				return err
			}
		}
		if len(out.LastEvaluatedKey) == 0 {
			return nil
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}
}

func (s *DynamoDBHeadScanner) StoreID(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.storeID != "" {
		return s.storeID, nil
	}
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"aid": &types.AttributeValueMemberS{Value: configAggregateID},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	value, ok := out.Item["store_id"].(*types.AttributeValueMemberS)
	if !ok {
		return "", fmt.Errorf("head configuration item %q has no store_id", configAggregateID)
	}
	s.storeID = value.Value
	return s.storeID, nil
}

func headFromItem(item map[string]types.AttributeValue) (Head, error) {
	aid, ok := item["aid"].(*types.AttributeValueMemberS)
	if !ok {
		return Head{}, fmt.Errorf("head item has no aid string")
	}
	typeName, ok := item["type_name"].(*types.AttributeValueMemberS)
	if !ok {
		return Head{}, fmt.Errorf("head item %s has no type_name string", aid.Value)
	}
	seqNr, ok := item["seq_nr"].(*types.AttributeValueMemberN)
	if !ok {
		return Head{}, fmt.Errorf("head item %s has no seq_nr number", aid.Value)
	}
	parsedSeqNr, err := strconv.ParseInt(seqNr.Value, 10, 64)
	if err != nil {
		return Head{}, err
	}
	return Head{AggregateID: aid.Value, TypeName: typeName.Value, SeqNr: parsedSeqNr}, nil
}
