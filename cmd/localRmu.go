package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"fmt"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"reflect"
	"time"
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
		headTableName := env.String("head", "STREAM_HEAD_TABLE_NAME")
		journalTableName := env.String("journal", "STREAM_JOURNAL_TABLE_NAME")
		streamMaxItemCount := env.Int64(100, "STREAM_MAX_ITEM_COUNT")

		slog.Info(fmt.Sprintf("AWS_REGION = %v", awsRegion))
		slog.Info(fmt.Sprintf("API_HOST = %v", apiHost))
		slog.Info(fmt.Sprintf("apiPort = %v", apiPort))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_ENDPOINT_URL = %v", awsDynamoDBEndpointUrl))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_ACCESS_KEY_ID = %v", awsDynamoDBAccessKeyId))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_SECRET_ACCESS_KEY = %v", awsDynamoDBSecretKey))
		slog.Info(fmt.Sprintf("STREAM_HEAD_TABLE_NAME = %v", headTableName))
		slog.Info(fmt.Sprintf("STREAM_JOURNAL_TABLE_NAME = %v", journalTableName))
		slog.Info(fmt.Sprintf("STREAM_MAX_ITEM_COUNT = %v", streamMaxItemCount))

		awsCfg, err := loadAWSConfig(awsRegion, awsDynamoDBEndpointUrl, awsDynamoDBAccessKeyId, awsDynamoDBSecretKey)
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
		journalReader := rmu.NewDynamoDBJournalReader(dynamodbClient, journalTableName)
		readModelUpdater := rmu.NewReadModelUpdater(dao, journalReader)

		for {
			err := streamDriver(context.Background(), dynamodbClient, dynamodbStreamsClient, headTableName, int32(streamMaxItemCount), &readModelUpdater)
			if err != nil {
				slog.Warn(fmt.Sprintf("An error has occurred, but stream processing is restarted. "+
					"If this error persists, the read model condition may be incorrect.: error = %v", err))
				time.Sleep(1 * time.Second)
				continue
			}
		}
	},
}

func loadAWSConfig(awsRegion, endpointUrl, accessKeyId, secretKey string) (aws.Config, error) {
	if endpointUrl == "" && accessKeyId == "" && secretKey == "" && awsRegion == "" {
		return config.LoadDefaultConfig(context.Background())
	}
	if awsRegion == "" {
		return aws.Config{}, fmt.Errorf("AWS_REGION is required")
	}
	if endpointUrl == "" {
		return aws.Config{}, fmt.Errorf("AWS_DYNAMODB_ENDPOINT_URL is required")
	}
	if accessKeyId == "" {
		return aws.Config{}, fmt.Errorf("AWS_DYNAMODB_ACCESS_KEY_ID is required")
	}
	if secretKey == "" {
		return aws.Config{}, fmt.Errorf("AWS_DYNAMODB_SECRET_ACCESS_KEY is required")
	}
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpointUrl,
				SigningRegion: region,
			}, nil
		})
	return config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretKey, "x")),
	)
}

// streamDriver polls every shard of the head table's stream. The first iterator
// starts at TRIM_HORIZON so a restart replays the retained records; the
// read_model_positions table drops what was already applied.
func streamDriver(ctx context.Context, dynamoDbClient *dynamodb.Client, dynamoDbStreamsClient *dynamodbstreams.Client, headTableName string, maxItemCount int32, readModelUpdater *rmu.ReadModelUpdater) error {
	describeTable, err := describeTable(dynamoDbClient, headTableName)
	if err != nil {
		return err
	}
	if describeTable.Table == nil || describeTable.Table.LatestStreamArn == nil {
		return fmt.Errorf("head table %q has no stream", headTableName)
	}
	streamArn := describeTable.Table.LatestStreamArn

	iterators := map[string]*string{}
	completedShards := map[string]bool{}
	for {
		var lastEvaluatedShardId *string
		for {
			describeStream, err := getDescribeStream(streamArn, lastEvaluatedShardId, dynamoDbStreamsClient)
			if err != nil {
				return err
			}
			for _, shard := range describeStream.StreamDescription.Shards {
				shardId := aws.ToString(shard.ShardId)
				if completedShards[shardId] {
					continue
				}
				iterator, ok := iterators[shardId]
				if !ok {
					getShardIterator, err := getShardIterator(dynamoDbStreamsClient, streamArn, shard)
					if err != nil {
						return err
					}
					iterator = getShardIterator.ShardIterator
				}
				for iterator != nil {
					getRecords, err := getRecords(ctx, dynamoDbStreamsClient, iterator, maxItemCount)
					if err != nil {
						return err
					}
					if err := updateRecords(ctx, getRecords, readModelUpdater, streamArn); err != nil {
						return err
					}
					if getRecords.NextShardIterator == nil {
						delete(iterators, shardId)
						completedShards[shardId] = true
						break
					}
					iterators[shardId] = getRecords.NextShardIterator
					iterator = getRecords.NextShardIterator
					if len(getRecords.Records) == 0 {
						break
					}
				}
			}
			if describeStream.StreamDescription.LastEvaluatedShardId == nil {
				break
			}
			lastEvaluatedShardId = describeStream.StreamDescription.LastEvaluatedShardId
		}
		time.Sleep(1 * time.Second)
	}
}

func describeTable(dynamoDbClient *dynamodb.Client, headTableName string) (*dynamodb.DescribeTableOutput, error) {
	return dynamoDbClient.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(headTableName),
	})
}

func updateRecords(ctx context.Context, getRecords *dynamodbstreams.GetRecordsOutput, readModelUpdater *rmu.ReadModelUpdater, streamArn *string) error {
	for _, record := range getRecords.Records {
		eventRecord, err := convertToEvent(record, streamArn)
		if err != nil {
			return err
		}
		if err := readModelUpdater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
			Records: []dynamodbevents.DynamoDBEventRecord{eventRecord},
		}); err != nil {
			return err
		}
	}
	return nil
}

func getRecords(ctx context.Context, dynamoDbStreamsClient *dynamodbstreams.Client, shardIterator *string, maxItemCount int32) (*dynamodbstreams.GetRecordsOutput, error) {
	return dynamoDbStreamsClient.GetRecords(ctx, &dynamodbstreams.GetRecordsInput{
		ShardIterator: shardIterator,
		Limit:         aws.Int32(maxItemCount),
	})
}

func getShardIterator(dynamoDbStreamsClient *dynamodbstreams.Client, streamArn *string, shard types.Shard) (*dynamodbstreams.GetShardIteratorOutput, error) {
	return dynamoDbStreamsClient.GetShardIterator(context.Background(), &dynamodbstreams.GetShardIteratorInput{
		StreamArn:         streamArn,
		ShardId:           shard.ShardId,
		ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
	})
}

func getDescribeStream(streamArn *string, exclusiveStartShardId *string, dynamoDbStreamsClient *dynamodbstreams.Client) (*dynamodbstreams.DescribeStreamOutput, error) {
	input := &dynamodbstreams.DescribeStreamInput{
		StreamArn: streamArn,
	}
	if exclusiveStartShardId != nil {
		input.ExclusiveStartShardId = exclusiveStartShardId
	}
	return dynamoDbStreamsClient.DescribeStream(context.Background(), input)
}

func convertToEvent(record types.Record, streamArn *string) (dynamodbevents.DynamoDBEventRecord, error) {
	if record.Dynamodb == nil {
		return dynamodbevents.DynamoDBEventRecord{}, fmt.Errorf("stream record %s has no dynamodb data", aws.ToString(record.EventID))
	}
	keys, err := convertAttributeMap(record.Dynamodb.Keys)
	if err != nil {
		return dynamodbevents.DynamoDBEventRecord{}, err
	}
	newImage, err := convertAttributeMap(record.Dynamodb.NewImage)
	if err != nil {
		return dynamodbevents.DynamoDBEventRecord{}, err
	}
	approximateCreationDateTime := time.Time{}
	if record.Dynamodb.ApproximateCreationDateTime != nil {
		approximateCreationDateTime = *record.Dynamodb.ApproximateCreationDateTime
	}
	return dynamodbevents.DynamoDBEventRecord{
		AWSRegion: aws.ToString(record.AwsRegion),
		Change: dynamodbevents.DynamoDBStreamRecord{
			ApproximateCreationDateTime: dynamodbevents.SecondsEpochTime{Time: approximateCreationDateTime},
			Keys:                        keys,
			NewImage:                    newImage,
			SequenceNumber:              aws.ToString(record.Dynamodb.SequenceNumber),
			SizeBytes:                   aws.ToInt64(record.Dynamodb.SizeBytes),
			StreamViewType:              string(record.Dynamodb.StreamViewType),
		},
		EventID:        aws.ToString(record.EventID),
		EventName:      string(record.EventName),
		EventSource:    aws.ToString(record.EventSource),
		EventVersion:   aws.ToString(record.EventVersion),
		EventSourceArn: aws.ToString(streamArn),
	}, nil
}

func convertTo(value types.AttributeValue) mo.Result[dynamodbevents.DynamoDBAttributeValue] {
	var av dynamodbevents.DynamoDBAttributeValue
	switch value := value.(type) {
	case *types.AttributeValueMemberNULL:
		av = dynamodbevents.NewNullAttribute()
	case *types.AttributeValueMemberS:
		av = dynamodbevents.NewStringAttribute(value.Value)
	case *types.AttributeValueMemberN:
		av = dynamodbevents.NewNumberAttribute(value.Value)
	case *types.AttributeValueMemberB:
		av = dynamodbevents.NewBinaryAttribute(value.Value)
	case *types.AttributeValueMemberBS:
		av = dynamodbevents.NewBinarySetAttribute(value.Value)
	case *types.AttributeValueMemberNS:
		av = dynamodbevents.NewNumberSetAttribute(value.Value)
	case *types.AttributeValueMemberSS:
		av = dynamodbevents.NewStringSetAttribute(value.Value)
	case *types.AttributeValueMemberL:
		l := make([]dynamodbevents.DynamoDBAttributeValue, 0, len(value.Value))
		for _, e := range value.Value {
			result := convertTo(e)
			if result.IsError() {
				return result
			}
			l = append(l, result.MustGet())
		}
		av = dynamodbevents.NewListAttribute(l)
	case *types.AttributeValueMemberM:
		m := make(map[string]dynamodbevents.DynamoDBAttributeValue, len(value.Value))
		for key, e := range value.Value {
			result := convertTo(e)
			if result.IsError() {
				return result
			}
			m[key] = result.MustGet()
		}
		av = dynamodbevents.NewMapAttribute(m)
	case *types.AttributeValueMemberBOOL:
		av = dynamodbevents.NewBooleanAttribute(value.Value)
	default:
		return mo.Err[dynamodbevents.DynamoDBAttributeValue](fmt.Errorf("unknown type: %s", reflect.TypeOf(value)))
	}
	return mo.Ok(av)
}

func convertAttributeMap(recordMap map[string]types.AttributeValue) (map[string]dynamodbevents.DynamoDBAttributeValue, error) {
	result := make(map[string]dynamodbevents.DynamoDBAttributeValue, len(recordMap))
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
