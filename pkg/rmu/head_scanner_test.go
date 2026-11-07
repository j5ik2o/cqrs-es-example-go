package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/test"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

func TestDynamoDBHeadScanner_ScansEveryPageAndReadsStoreID(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateLocalStackContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })

	client, err := test.CreateDynamoDBClient(t, ctx, container)
	require.NoError(t, err)
	test.WaitForDynamoDB(t, ctx, client)
	require.NoError(t, test.CreateHeadTable(t, ctx, client, "head"))

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("head"),
		Item: map[string]types.AttributeValue{
			"aid":            &types.AttributeValueMemberS{Value: configAggregateID},
			"store_id":       &types.AttributeValueMemberS{Value: "store-1"},
			"layout_version": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	require.NoError(t, err)

	const aggregateCount = 5
	aggregateIDs := make([]string, 0, aggregateCount)
	for i := 0; i < aggregateCount; i++ {
		groupChatID := models.NewGroupChatId()
		aggregateID := groupChatID.AsString()
		aggregateIDs = append(aggregateIDs, aggregateID)
		_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("head"),
			Item: map[string]types.AttributeValue{
				"aid":       &types.AttributeValueMemberS{Value: aggregateID},
				"type_name": &types.AttributeValueMemberS{Value: models.GroupChatIdPrefix},
				"seq_nr":    &types.AttributeValueMemberN{Value: strconv.Itoa(i + 1)},
			},
		})
		require.NoError(t, err)
	}

	// One item per page forces the LastEvaluatedKey loop.
	scanner := NewDynamoDBHeadScanner(client, "head", WithPageLimit(1))
	var heads []Head
	require.NoError(t, scanner.ScanHeads(ctx, func(head Head) error {
		heads = append(heads, head)
		return nil
	}))
	require.Len(t, heads, aggregateCount)
	scanned := map[string]bool{}
	for _, head := range heads {
		require.NotEqual(t, configAggregateID, head.AggregateID)
		require.Equal(t, models.GroupChatIdPrefix, head.TypeName)
		scanned[head.AggregateID] = true
	}
	for _, aggregateID := range aggregateIDs {
		require.True(t, scanned[aggregateID])
	}

	storeID, err := scanner.StoreID(ctx)
	require.NoError(t, err)
	require.Equal(t, "store-1", storeID)
}
