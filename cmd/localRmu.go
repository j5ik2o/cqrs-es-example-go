package cmd

import (
	"context"
	"cqrs-es-example-go/pkg/rmu"
	"errors"
	"fmt"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	streamtypes "github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
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

// errStreamTrimmed means the reader fell behind the stream retention window.
// The caller answers with a full resync (DY-15), not an iterator restart.
var errStreamTrimmed = errors.New("head stream records were trimmed")

// isExpiredIterator reports the 15-minute iterator expiry. The record is still
// in the stream, so the reader resumes after the last processed sequence number
// instead of resyncing everything.
func isExpiredIterator(err error) bool {
	var expired *streamtypes.ExpiredIteratorException
	return errors.As(err, &expired)
}

// isTrimmedData reports that the requested position is older than the 24-hour
// retention window. The reader must reconcile the whole head table.
func isTrimmedData(err error) bool {
	var trimmed *streamtypes.TrimmedDataAccessException
	return errors.As(err, &trimmed)
}

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
		maxEventsPerTransaction := env.Int(1, "STREAM_MAX_EVENTS_PER_TRANSACTION")
		recoveryTaskLimit := rmu.NormalizeRecoveryTaskLimit(
			env.Int(rmu.DefaultRecoveryTaskLimit, "RECOVERY_MAX_TASKS_PER_PASS"))

		retryPolicy := rmu.DefaultRetryPolicy()
		retryPolicy.MaxAttempts = env.Int(retryPolicy.MaxAttempts, "STREAM_RETRY_MAX_ATTEMPTS")
		retryPolicy.MaxElapsed = durationEnv("STREAM_RETRY_MAX_ELAPSED", retryPolicy.MaxElapsed)
		retryPolicy.InitialDelay = durationEnv("STREAM_RETRY_INITIAL_DELAY", 500*time.Millisecond)
		retryPolicy.MaxDelay = durationEnv("STREAM_RETRY_MAX_DELAY", 30*time.Second)

		slog.Info(fmt.Sprintf("AWS_REGION = %v", awsRegion))
		slog.Info(fmt.Sprintf("API_HOST = %v", apiHost))
		slog.Info(fmt.Sprintf("apiPort = %v", apiPort))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_ENDPOINT_URL = %v", awsDynamoDBEndpointUrl))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_ACCESS_KEY_ID = %v", awsDynamoDBAccessKeyId))
		slog.Info(fmt.Sprintf("AWS_DYNAMODB_SECRET_ACCESS_KEY = %v", awsDynamoDBSecretKey))
		slog.Info(fmt.Sprintf("STREAM_HEAD_TABLE_NAME = %v", headTableName))
		slog.Info(fmt.Sprintf("STREAM_JOURNAL_TABLE_NAME = %v", journalTableName))
		slog.Info(fmt.Sprintf("STREAM_MAX_ITEM_COUNT = %v", streamMaxItemCount))
		slog.Info(fmt.Sprintf("STREAM_MAX_EVENTS_PER_TRANSACTION = %v", maxEventsPerTransaction))
		slog.Info(fmt.Sprintf("STREAM_RETRY_MAX_ATTEMPTS = %v", retryPolicy.MaxAttempts))
		slog.Info(fmt.Sprintf("STREAM_RETRY_MAX_ELAPSED = %v", retryPolicy.MaxElapsed))
		slog.Info(fmt.Sprintf("RECOVERY_MAX_TASKS_PER_PASS = %v", recoveryTaskLimit))

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
		headScanner := rmu.NewDynamoDBHeadScanner(dynamodbClient, headTableName)
		applier := rmu.NewApplier(dao, journalReader, maxEventsPerTransaction)
		recoveryService := rmu.NewRecoveryService(
			rmu.NewMySQLRecoveryTaskStore(db), headScanner, applier,
			rmu.WithRecoveryTaskLimit(recoveryTaskLimit))
		readModelUpdater := rmu.NewReadModelUpdater(dao, journalReader,
			rmu.WithApplier(applier),
			rmu.WithRecoveryService(recoveryService),
			rmu.WithRetryPolicy(retryPolicy))
		resyncer := rmu.NewResyncer(applier, headScanner, recoveryService)

		ctx := context.Background()
		// The first start has no positions at all, and a restart may follow a
		// stop longer than the 24-hour stream retention. Reconcile every head
		// before consuming the stream; the stream then covers the gap between
		// the scan and the first read. A transient scan failure is retried
		// rather than left until the next trim event.
		resyncBackoff := rmu.NewBackoff(500*time.Millisecond, 30*time.Second)
		resyncUntilResolved(ctx, resyncer, retryPolicy, resyncBackoff, "initial reconciliation")
		if result, err := recoveryService.RecoverPending(ctx); err != nil {
			slog.Warn("some pending recovery tasks did not complete",
				"completed", result.Completed, "pending", result.Pending, "error", err)
		} else if result.Completed > 0 {
			slog.Info("completed pending recovery tasks",
				"count", result.Completed, "pending", result.Pending, "deferred", result.Deferred)
		}

		backoff := rmu.NewBackoff(500*time.Millisecond, 30*time.Second)
		pollBackoff := rmu.NewBackoff(250*time.Millisecond, 5*time.Second)
		for {
			err := streamDriver(ctx, dynamodbClient, dynamodbStreamsClient, headTableName,
				int32(streamMaxItemCount), &readModelUpdater, pollBackoff)
			if errors.Is(err, errStreamTrimmed) {
				slog.Warn("the reader fell behind the stream retention window; running a full resync", "error", err)
				resyncUntilResolved(ctx, resyncer, retryPolicy, resyncBackoff, "resync after trimmed stream")
				backoff.Reset()
				continue
			}
			slog.Warn(fmt.Sprintf("An error has occurred, but stream processing is restarted. "+
				"If this error persists, the read model condition may be incorrect.: error = %v", err))
			if sleepErr := backoffSleep(ctx, backoff.Next()); sleepErr != nil {
				return
			}
		}
	},
}

func resyncUntilResolved(ctx context.Context, resyncer *rmu.Resyncer, policy rmu.RetryPolicy, backoff *rmu.Backoff, label string) {
	for {
		err := resync(ctx, resyncer, policy, label)
		if err == nil {
			return
		}
		if !rmu.IsRetryable(err) {
			// Permanent per-aggregate failures were recorded as recovery tasks.
			// Continuing keeps the stream moving while they are repaired.
			return
		}
		if sleepErr := backoffSleep(ctx, backoff.Next()); sleepErr != nil {
			return
		}
	}
}

func resync(ctx context.Context, resyncer *rmu.Resyncer, policy rmu.RetryPolicy, label string) error {
	err := policy.Do(ctx, func() error {
		result, err := resyncer.Resync(ctx)
		if err == nil {
			slog.Info(label+" finished", "heads", result.Heads, "advanced", result.Advanced, "skipped", result.Skipped)
		}
		return err
	})
	if err != nil {
		slog.Error(label+" did not complete", "error", err)
	}
	return err
}

func durationEnv(name string, fallback time.Duration) time.Duration {
	raw := env.String("", name)
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		panic(fmt.Sprintf("%s must be a duration: %v", name, err))
	}
	return parsed
}

func backoffSleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
// read_model_positions table drops what was already applied. An expired
// iterator is refreshed from the last processed sequence number, which is not
// the same as data loss.
func streamDriver(ctx context.Context, dynamoDbClient *dynamodb.Client, dynamoDbStreamsClient *dynamodbstreams.Client, headTableName string, maxItemCount int32, readModelUpdater *rmu.ReadModelUpdater, pollBackoff *rmu.Backoff) error {
	describeTable, err := describeTable(ctx, dynamoDbClient, headTableName)
	if err != nil {
		return err
	}
	if describeTable.Table == nil || describeTable.Table.LatestStreamArn == nil {
		return fmt.Errorf("head table %q has no stream", headTableName)
	}
	streamArn := describeTable.Table.LatestStreamArn

	iterators := map[string]*string{}
	lastSequenceNumbers := map[string]string{}
	blockedAggregates := map[string]bool{}
	completedShards := map[string]bool{}
	for {
		var lastEvaluatedShardId *string
		for {
			describeStream, err := getDescribeStream(ctx, streamArn, lastEvaluatedShardId, dynamoDbStreamsClient)
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
					iterator, err = nextShardIterator(ctx, dynamoDbStreamsClient, streamArn, shard, lastSequenceNumbers[shardId])
					if err != nil {
						return err
					}
				}
				for iterator != nil {
					getRecords, err := getRecords(ctx, dynamoDbStreamsClient, iterator, maxItemCount)
					if err != nil {
						if isExpiredIterator(err) {
							slog.Warn("stream iterator expired; resuming after the last processed sequence number",
								"shard", shardId, "sequence_number", lastSequenceNumbers[shardId])
							iterator, err = nextShardIterator(ctx, dynamoDbStreamsClient, streamArn, shard, lastSequenceNumbers[shardId])
							if err != nil {
								return err
							}
							continue
						}
						if isTrimmedData(err) {
							return fmt.Errorf("%w: shard %s: %v", errStreamTrimmed, shardId, err)
						}
						return err
					}
					if err := updateRecords(ctx, getRecords, readModelUpdater, streamArn, lastSequenceNumbers, blockedAggregates, shardId); err != nil {
						return err
					}
					if len(getRecords.Records) > 0 {
						pollBackoff.Reset()
					}
					if getRecords.NextShardIterator == nil {
						delete(iterators, shardId)
						completedShards[shardId] = true
						break
					}
					iterators[shardId] = getRecords.NextShardIterator
					iterator = getRecords.NextShardIterator
					if len(getRecords.Records) == 0 {
						if err := backoffSleep(ctx, pollBackoff.Next()); err != nil {
							return err
						}
						break
					}
				}
			}
			if describeStream.StreamDescription.LastEvaluatedShardId == nil {
				break
			}
			lastEvaluatedShardId = describeStream.StreamDescription.LastEvaluatedShardId
		}
		if err := backoffSleep(ctx, pollBackoff.Next()); err != nil {
			return err
		}
	}
}

func describeTable(ctx context.Context, dynamoDbClient *dynamodb.Client, headTableName string) (*dynamodb.DescribeTableOutput, error) {
	return dynamoDbClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(headTableName),
	})
}

// updateRecords applies one poll's records. A permanent failure is parked as a
// recovery task and only blocks that aggregate; other aggregates keep moving.
// A transient failure returns to the caller so the reader backs off.
func updateRecords(ctx context.Context, getRecords *dynamodbstreams.GetRecordsOutput, readModelUpdater *rmu.ReadModelUpdater, streamArn *string, lastSequenceNumbers map[string]string, blockedAggregates map[string]bool, shardId string) error {
	for _, record := range getRecords.Records {
		eventRecord, err := convertToEvent(record, streamArn)
		if err != nil {
			return err
		}
		aggregateID, hasAggregateID := rmu.RecordAggregateID(eventRecord)
		if hasAggregateID && blockedAggregates[aggregateID] {
			slog.Warn("skip head record of an aggregate with an unapplied event", "aid", aggregateID)
			if record.Dynamodb != nil {
				lastSequenceNumbers[shardId] = aws.ToString(record.Dynamodb.SequenceNumber)
			}
			continue
		}
		if err := readModelUpdater.UpdateReadModel(ctx, dynamodbevents.DynamoDBEvent{
			Records: []dynamodbevents.DynamoDBEventRecord{eventRecord},
		}); err != nil {
			if _, isPermanent := rmu.ClassifyFailure(err); isPermanent {
				if hasAggregateID {
					blockedAggregates[aggregateID] = true
				}
				slog.Error("head record failed permanently; recorded a recovery task",
					"aid", aggregateID, "event_id", eventRecord.EventID, "error", err)
				if record.Dynamodb != nil {
					lastSequenceNumbers[shardId] = aws.ToString(record.Dynamodb.SequenceNumber)
				}
				continue
			}
			return err
		}
		if record.Dynamodb != nil {
			lastSequenceNumbers[shardId] = aws.ToString(record.Dynamodb.SequenceNumber)
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

func nextShardIterator(ctx context.Context, dynamoDbStreamsClient *dynamodbstreams.Client, streamArn *string, shard streamtypes.Shard, lastSequenceNumber string) (*string, error) {
	input := &dynamodbstreams.GetShardIteratorInput{
		StreamArn:         streamArn,
		ShardId:           shard.ShardId,
		ShardIteratorType: streamtypes.ShardIteratorTypeTrimHorizon,
	}
	if lastSequenceNumber != "" {
		input.ShardIteratorType = streamtypes.ShardIteratorTypeAfterSequenceNumber
		input.SequenceNumber = aws.String(lastSequenceNumber)
	}
	out, err := dynamoDbStreamsClient.GetShardIterator(ctx, input)
	if err != nil {
		return nil, err
	}
	return out.ShardIterator, nil
}

func getDescribeStream(ctx context.Context, streamArn *string, exclusiveStartShardId *string, dynamoDbStreamsClient *dynamodbstreams.Client) (*dynamodbstreams.DescribeStreamOutput, error) {
	input := &dynamodbstreams.DescribeStreamInput{
		StreamArn: streamArn,
	}
	if exclusiveStartShardId != nil {
		input.ExclusiveStartShardId = exclusiveStartShardId
	}
	return dynamoDbStreamsClient.DescribeStream(ctx, input)
}

func convertToEvent(record streamtypes.Record, streamArn *string) (dynamodbevents.DynamoDBEventRecord, error) {
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

func convertTo(value streamtypes.AttributeValue) mo.Result[dynamodbevents.DynamoDBAttributeValue] {
	var av dynamodbevents.DynamoDBAttributeValue
	switch value := value.(type) {
	case *streamtypes.AttributeValueMemberNULL:
		av = dynamodbevents.NewNullAttribute()
	case *streamtypes.AttributeValueMemberS:
		av = dynamodbevents.NewStringAttribute(value.Value)
	case *streamtypes.AttributeValueMemberN:
		av = dynamodbevents.NewNumberAttribute(value.Value)
	case *streamtypes.AttributeValueMemberB:
		av = dynamodbevents.NewBinaryAttribute(value.Value)
	case *streamtypes.AttributeValueMemberBS:
		av = dynamodbevents.NewBinarySetAttribute(value.Value)
	case *streamtypes.AttributeValueMemberNS:
		av = dynamodbevents.NewNumberSetAttribute(value.Value)
	case *streamtypes.AttributeValueMemberSS:
		av = dynamodbevents.NewStringSetAttribute(value.Value)
	case *streamtypes.AttributeValueMemberL:
		l := make([]dynamodbevents.DynamoDBAttributeValue, 0, len(value.Value))
		for _, e := range value.Value {
			result := convertTo(e)
			if result.IsError() {
				return result
			}
			l = append(l, result.MustGet())
		}
		av = dynamodbevents.NewListAttribute(l)
	case *streamtypes.AttributeValueMemberM:
		m := make(map[string]dynamodbevents.DynamoDBAttributeValue, len(value.Value))
		for key, e := range value.Value {
			result := convertTo(e)
			if result.IsError() {
				return result
			}
			m[key] = result.MustGet()
		}
		av = dynamodbevents.NewMapAttribute(m)
	case *streamtypes.AttributeValueMemberBOOL:
		av = dynamodbevents.NewBooleanAttribute(value.Value)
	default:
		return mo.Err[dynamodbevents.DynamoDBAttributeValue](fmt.Errorf("unknown type: %s", reflect.TypeOf(value)))
	}
	return mo.Ok(av)
}

func convertAttributeMap(recordMap map[string]streamtypes.AttributeValue) (map[string]dynamodbevents.DynamoDBAttributeValue, error) {
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
