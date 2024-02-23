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
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/samber/mo"
	"log/slog"
	"os"
	"reflect"
	"time"

	"github.com/spf13/cobra"
)

const rmuDefaultPort = 28081

// localRmuCmd represents the localRmu command
var localRmuCmd = &cobra.Command{
	Use:   "localRmu",
	Short: "Read Model Updater for Local",
	Long:  "Read Model Updater for Local",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		awsRegion := env.String("", "AWS_REGION")
		apiHost := env.String("0.0.0.0", "API_HOST")
		apiPort := env.Int(rmuDefaultPort, "API_PORT")
		awsDynamoDBEndpointUrl := env.String("", "AWS_DYNAMODB_ENDPOINT_URL")
		awsDynamoDBAccessKeyId := env.String("", "AWS_DYNAMODB_ACCESS_KEY_ID")
		awsDynamoDBSecretKey := env.String("", "AWS_DYNAMODB_SECRET_ACCESS_KEY")
		streamJournalTableName := env.String("journal", "STREAM_JOURNAL_TABLE_NAME")
		streamMaxItemCount := env.Int64(100, "STREAM_MAX_ITEM_COUNT")

		slog.Info(fmt.Sprintf("awsRegion = %v", awsRegion))
		slog.Info(fmt.Sprintf("apiHost = %v", apiHost))
		slog.Info(fmt.Sprintf("apiPort = %v", apiPort))
		slog.Info(fmt.Sprintf("awsDynamoDBEndpointUrl = %v", awsDynamoDBEndpointUrl))
		slog.Info(fmt.Sprintf("awsDynamoDBAccessKeyId = %v", awsDynamoDBAccessKeyId))
		slog.Info(fmt.Sprintf("awsDynamoDBSecretKey = %v", awsDynamoDBSecretKey))
		slog.Info(fmt.Sprintf("streamJournalTableName = %v", streamJournalTableName))
		slog.Info(fmt.Sprintf("streamMaxItemCount = %v", streamMaxItemCount))

		var awsCfg aws.Config
		var err error
		if awsDynamoDBEndpointUrl == "" && awsDynamoDBAccessKeyId == "" && awsDynamoDBSecretKey == "" && awsRegion == "" {
			awsCfg, err = config.LoadDefaultConfig(context.Background())
		} else {
			if awsRegion == "" {
				panic("AWS_REGION is required")
			}
			if awsDynamoDBEndpointUrl == "" {
				panic("AWS_DYNAMODB_ENDPOINT_URL is required")
			}
			if awsDynamoDBAccessKeyId == "" {
				panic("AWS_DYNAMODB_ACCESS_KEY_ID is required")
			}
			if awsDynamoDBSecretKey == "" {
				panic("AWS_DYNAMODB_SECRET_ACCESS_KEY is required")
			}
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

		dbUrl := env.String("", "DATABASE_URL")
		if dbUrl == "" {
			panic("DATABASE_URL is required")
		}
		dataSourceName := fmt.Sprintf("%s?parseTime=true", dbUrl)
		db, err := sqlx.Connect("mysql", dataSourceName)
		if err != nil {
			panic(err.Error())
		}
		defer func(db *sqlx.DB) {
			if db != nil {
				err := db.Close()
				if err != nil {
					panic(err.Error())
				}
			}
		}(db)
		dao := rmu.NewGroupChatDaoImpl(db)
		readModelUpdater := rmu.NewReadModelUpdater(&dao)

		for {
			err := streamDriver(dynamodbClient, dynamodbStreamsClient, streamJournalTableName, streamMaxItemCount, &readModelUpdater)
			if err != nil {
				slog.Warn(fmt.Sprintf("An error has occurred, but stream processing is restarted. "+
					"If this error persists, the read model condition may be incorrect.: error = %v", err))
				time.Sleep(1 * time.Second)
				continue
			}
		}
	},
}

func streamDriver(dynamoDbClient *dynamodb.Client, dynamoDbStreamsClient *dynamodbstreams.Client, streamJournalTableName string, maxItemCount int64, readModelUpdater *rmu.ReadModelUpdater) error {
	describeTable, err := describeTable(dynamoDbClient, streamJournalTableName)
	if err != nil {
		return err
	}
	streamArn := describeTable.Table.LatestStreamArn
	lastEvaluatedShardId := ""

	for {
		slog.Info(fmt.Sprintf("streamArn = %s", *streamArn))
		slog.Info(fmt.Sprintf("maxItemCount = %d", maxItemCount))

		describeStream, err := getDescribeStream(streamArn, lastEvaluatedShardId, dynamoDbStreamsClient)
		if err != nil {
			return err
		}

		for _, shard := range describeStream.StreamDescription.Shards {
			getShardIterator, err := getShardIterator(dynamoDbStreamsClient, streamArn, shard)
			if err != nil {
				return err
			}
			shardIterator := getShardIterator.ShardIterator
			processedRecordCount := 0
			for shardIterator != nil && processedRecordCount < int(maxItemCount) {
				getRecords, err := getRecords(dynamoDbStreamsClient, shardIterator)
				if err != nil {
					return err
				}
				err = updateRecords(getRecords, readModelUpdater, streamArn)
				if err != nil {
					return err
				}
				processedRecordCount += len(getRecords.Records)
				shardIterator = getRecords.NextShardIterator
			}
		}

		if describeStream.StreamDescription.LastEvaluatedShardId == nil {
			break
		}
		lastEvaluatedShardId = *describeStream.StreamDescription.LastEvaluatedShardId
	}
	return nil
}

func describeTable(dynamoDbClient *dynamodb.Client, journalTableName string) (*dynamodb.DescribeTableOutput, error) {
	describeTableResult, err := dynamoDbClient.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(journalTableName),
	})
	return describeTableResult, err
}

func updateRecords(getRecords *dynamodbstreams.GetRecordsOutput, readModelUpdater *rmu.ReadModelUpdater, streamArn *string) error {
	for _, record := range getRecords.Records {
		keys, err := getKeys(record)
		if err != nil {
			return err
		}
		item := getItem(record, err)
		if err != nil {
			return err
		}
		err = readModelUpdater.UpdateReadModel(context.Background(), convertToEvent(record, keys, item, streamArn))
		if err != nil {
			return err
		}
	}
	return nil
}

func getItem(record types.Record, err error) map[string]events.DynamoDBAttributeValue {
	itemMap := record.Dynamodb.NewImage
	newItem, err := convertAttributeMap(itemMap)
	return newItem
}

func getKeys(record types.Record) (map[string]events.DynamoDBAttributeValue, error) {
	keysMap := record.Dynamodb.Keys
	keys, err := convertAttributeMap(keysMap)
	return keys, err
}

func getRecords(dynamoDbStreamsClient *dynamodbstreams.Client, shardIterator *string) (*dynamodbstreams.GetRecordsOutput, error) {
	getRecordsResult, err := dynamoDbStreamsClient.GetRecords(context.Background(), &dynamodbstreams.GetRecordsInput{
		ShardIterator: shardIterator,
	})
	return getRecordsResult, err
}

func getShardIterator(dynamoDbStreamsClient *dynamodbstreams.Client, streamArn *string, shard types.Shard) (*dynamodbstreams.GetShardIteratorOutput, error) {
	getShardIterator, err := dynamoDbStreamsClient.GetShardIterator(context.Background(), &dynamodbstreams.GetShardIteratorInput{
		StreamArn:         streamArn,
		ShardId:           shard.ShardId,
		ShardIteratorType: types.ShardIteratorTypeLatest,
	})
	return getShardIterator, err
}

func getDescribeStream(streamArn *string, lastEvaluatedShardId string, dynamoDbStreamsClient *dynamodbstreams.Client) (*dynamodbstreams.DescribeStreamOutput, error) {
	describeStreamRequest := &dynamodbstreams.DescribeStreamInput{
		StreamArn: streamArn,
	}
	if lastEvaluatedShardId != "" {
		describeStreamRequest.ExclusiveStartShardId = &lastEvaluatedShardId
	}
	describeStreamResponse, err := dynamoDbStreamsClient.DescribeStream(context.Background(), describeStreamRequest)
	if err != nil {
		return nil, err
	}
	return describeStreamResponse, nil
}

func convertToEvent(record types.Record, keys map[string]events.DynamoDBAttributeValue, newItem map[string]events.DynamoDBAttributeValue, streamArn *string) events.DynamoDBEvent {
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
