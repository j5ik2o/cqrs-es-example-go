package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
	"github.com/olivere/env"
	"github.com/samber/mo"
	"reflect"
	"time"

	"github.com/spf13/cobra"
)

// localRmuCmd represents the localRmu command
var localRmuCmd = &cobra.Command{
	Use:   "localRmu",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		awsRegion := env.String("", "AWS_REGION")
		apiHost := env.String("0.0.0.0", "API_HOST")
		apiPort := env.Int(8080, "API_PORT")
		journalTableName := env.String("journal", "PERSISTENCE_JOURNAL_TABLE_NAME")
		snapshotTableName := env.String("snapshot", "PERSISTENCE_SNAPSHOT_TABLE_NAME")
		journalAidIndexName := env.String("journal-aid-index", "PERSISTENCE_JOURNAL_AID_INDEX_NAME")
		snapshotAidIndexName := env.String("snapshot-aid-index", "PERSISTENCE_SNAPSHOT_AID_INDEX_NAME")
		shardCount := env.Int64(10, "PERSISTENCE_SHARD_COUNT")
		awsDynamoDBEndpointUrl := env.String("", "AWS_DYNAMODB_ENDPOINT_URL")
		awsDynamoDBAccessKeyId := env.String("", "AWS_DYNAMODB_ACCESS_KEY_ID")
		awsDynamoDBSecretKey := env.String("", "AWS_DYNAMODB_SECRET_ACCESS_KEY")
		streamJournalTableName := env.String("journal", "STREAM_JOURNAL_TABLE_NAME")
		streamMaxItemCount := env.Int64(100, "STREAM_MAX_ITEM_COUNT")

		fmt.Printf("awsRegion = %v\n", awsRegion)
		fmt.Printf("apiHost = %v\n", apiHost)
		fmt.Printf("apiPort = %v\n", apiPort)
		fmt.Printf("journalTableName = %v\n", journalTableName)
		fmt.Printf("snapshotTableName = %v\n", snapshotTableName)
		fmt.Printf("journalAidIndexName = %v\n", journalAidIndexName)
		fmt.Printf("snapshotAidIndexName = %v\n", snapshotAidIndexName)
		fmt.Printf("shardCount = %v\n", shardCount)
		fmt.Printf("awsDynamoDBEndpointUrl = %v\n", awsDynamoDBEndpointUrl)
		fmt.Printf("awsDynamoDBAccessKeyId = %v\n", awsDynamoDBAccessKeyId)
		fmt.Printf("awsDynamoDBSecretKey = %v\n", awsDynamoDBSecretKey)
		fmt.Printf("streamJournalTableName = %v\n", streamJournalTableName)
		fmt.Printf("streamMaxItemCount = %v\n", streamMaxItemCount)

		var awsCfg aws.Config
		var err error
		if awsDynamoDBEndpointUrl == "" && awsDynamoDBAccessKeyId == "" && awsDynamoDBSecretKey == "" && awsRegion == "" {
			awsCfg, err = config.LoadDefaultConfig(context.Background())
		} else {
			customResolver := aws.EndpointResolverWithOptionsFunc(
				func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           awsDynamoDBEndpointUrl,
						SigningRegion: region,
					}, nil
				})
			awsCfg, err = config.LoadDefaultConfig(context.Background(),
				config.WithRegion(awsRegion),
				config.WithEndpointResolverWithOptions(customResolver),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsDynamoDBAccessKeyId, awsDynamoDBSecretKey, "x")),
			)
		}
		if err != nil {
			panic(err)
		}

		dynamodbClient := dynamodb.NewFromConfig(awsCfg)
		dynamodbStreamsClient := dynamodbstreams.NewFromConfig(awsCfg)

		for {
			err := streamDriver(dynamodbClient, dynamodbStreamsClient, journalTableName, streamMaxItemCount)
			if err != nil {
				fmt.Printf("An error has occurred, but stream processing is restarted. "+
					"If this error persists, the read model condition may be incorrect.: error = %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}
		}
	},
}

func streamDriver(dynamoDbClient *dynamodb.Client, dynamoDbStreamsClient *dynamodbstreams.Client, journalTableName string, maxItemCount int64) error {
	describeTableResult, err := dynamoDbClient.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(journalTableName),
	})
	if err != nil {
		return err
	}
	streamArn := describeTableResult.Table.LatestStreamArn
	lastEvaluatedShardId := ""
	for {
		fmt.Printf("streamArn = %s\n", *streamArn)
		fmt.Printf("maxItemCount = %d\n", maxItemCount)

		request := &dynamodbstreams.DescribeStreamInput{
			StreamArn: streamArn,
		}
		if lastEvaluatedShardId != "" {
			request.ExclusiveStartShardId = &lastEvaluatedShardId
		}
		describeStreamResult, err := dynamoDbStreamsClient.DescribeStream(context.Background(), request)
		if err != nil {
			return err
		}

		for _, shard := range describeStreamResult.StreamDescription.Shards {
			fmt.Printf("shard = %v\n", shard)
			getShardIterator, err := dynamoDbStreamsClient.GetShardIterator(context.Background(), &dynamodbstreams.GetShardIteratorInput{
				StreamArn:         streamArn,
				ShardId:           shard.ShardId,
				ShardIteratorType: types.ShardIteratorTypeLatest,
			})
			if err != nil {
				return err
			}
			shardIterator := getShardIterator.ShardIterator
			processedRecordCount := 0
			for shardIterator != nil && processedRecordCount < int(maxItemCount) {
				// fmt.Printf("shardIterator = %v\n", shardIterator)
				getRecordsResult, err := dynamoDbStreamsClient.GetRecords(context.Background(), &dynamodbstreams.GetRecordsInput{
					ShardIterator: shardIterator,
				})
				if err != nil {
					return err
				}
				for _, record := range getRecordsResult.Records {
					keysMap := record.Dynamodb.Keys
					keys, err := convertAttributeMap(keysMap)
					if err != nil {
						return err
					}

					itemMap := record.Dynamodb.NewImage
					newItem, err := convertAttributeMap(itemMap)
					if err != nil {
						return err
					}

					event := convertEvent(record, keys, newItem, streamArn)
					rmu.UpdateReadModel(context.Background(), event)
				}
				processedRecordCount += len(getRecordsResult.Records)
				shardIterator = getRecordsResult.NextShardIterator
			}
		}
		if describeStreamResult.StreamDescription.LastEvaluatedShardId == nil {
			break
		}
		lastEvaluatedShardId = *describeStreamResult.StreamDescription.LastEvaluatedShardId
	}
	return nil
}

func convertEvent(record types.Record, keys map[string]events.DynamoDBAttributeValue, newItem map[string]events.DynamoDBAttributeValue, streamArn *string) events.DynamoDBEvent {
	event := events.DynamoDBEvent{
		Records: []events.DynamoDBEventRecord{
			{
				AWSRegion: *record.AwsRegion,
				Change: events.DynamoDBStreamRecord{
					ApproximateCreationDateTime: events.SecondsEpochTime{Time: *record.Dynamodb.ApproximateCreationDateTime},
					Keys:                        keys,
					NewImage:                    newItem,
					SequenceNumber:              *record.Dynamodb.SequenceNumber,
					SizeBytes:                   *record.Dynamodb.SizeBytes,
					StreamViewType:              string(types.StreamViewTypeNewImage),
				},
				EventID:        *record.EventID,
				EventName:      string(types.OperationTypeInsert),
				EventSource:    *record.EventSource,
				EventVersion:   *record.EventVersion,
				EventSourceArn: *streamArn,
				UserIdentity: &events.DynamoDBUserIdentity{
					Type:        "Service",
					PrincipalID: "dynamodb.amazonaws.com",
				},
			},
		},
	}
	return event
}

func convertTo(value types.AttributeValue) mo.Result[events.DynamoDBAttributeValue] {
	var av events.DynamoDBAttributeValue
	switch value.(type) {
	case *types.AttributeValueMemberNULL:
		av = events.NewNullAttribute()
	case *types.AttributeValueMemberS:
		av = events.NewStringAttribute(value.(*types.AttributeValueMemberS).Value)
	case *types.AttributeValueMemberN:
		av = events.NewNumberAttribute(value.(*types.AttributeValueMemberN).Value)
	case *types.AttributeValueMemberB:
		av = events.NewBinaryAttribute(value.(*types.AttributeValueMemberB).Value)
	case *types.AttributeValueMemberBS:
		av = events.NewBinarySetAttribute(value.(*types.AttributeValueMemberBS).Value)
	case *types.AttributeValueMemberNS:
		av = events.NewNumberSetAttribute(value.(*types.AttributeValueMemberNS).Value)
	case *types.AttributeValueMemberSS:
		av = events.NewStringSetAttribute(value.(*types.AttributeValueMemberSS).Value)
	case *types.AttributeValueMemberL:
		var l []events.DynamoDBAttributeValue
		for _, e := range value.(*types.AttributeValueMemberL).Value {
			result := convertTo(e)
			if result.IsError() {
				return result
			}
			l = append(l, result.MustGet())
		}
		av = events.NewListAttribute(l)
	case *types.AttributeValueMemberM:
		m := make(map[string]events.DynamoDBAttributeValue)
		for key, e := range value.(*types.AttributeValueMemberM).Value {
			result := convertTo(e)
			if result.IsError() {
				return result
			}
			m[key] = result.MustGet()
		}
		av = events.NewMapAttribute(m)
	case *types.AttributeValueMemberBOOL:
		av = events.NewBooleanAttribute(value.(*types.AttributeValueMemberBOOL).Value)
	default:
		return mo.Err[events.DynamoDBAttributeValue](fmt.Errorf("unknown type: %s", reflect.TypeOf(value)))
	}
	return mo.Ok(av)
}

func convertAttributeMap(recordMap map[string]types.AttributeValue) (map[string]events.DynamoDBAttributeValue, error) {
	result := make(map[string]events.DynamoDBAttributeValue)
	for key, value := range recordMap {
		v, err := convertTo(value).Get()
		if err != nil {
			return nil, err
		}
		result[key] = v
	}
	return result, nil
}

func init() {
	rootCmd.AddCommand(localRmuCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// localRmuCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// localRmuCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
