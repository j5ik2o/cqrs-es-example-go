package test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	testcontainermysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

func CreateMySQLContainer(ctx context.Context) (*testcontainermysql.MySQLContainer, error) {
	container, err := testcontainermysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.2"),
		// mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		testcontainermysql.WithDatabase("ceer"),
		testcontainermysql.WithUsername("ceer"),
		testcontainermysql.WithPassword("ceer"),
		// testcontainermysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	return container, err
}

func CreateLocalStackContainer(ctx context.Context) (*localstack.LocalStackContainer, error) {
	container, err := localstack.RunContainer(
		ctx,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:2.1.0",
				Env: map[string]string{
					"SERVICES":              "dynamodb,dynamodbstreams",
					"DEFAULT_REGION":        "us-east-1",
					"EAGER_SERVICE_LOADING": "1",
					"DYNAMODB_SHARED_DB":    "1",
					"DYNAMODB_IN_MEMORY":    "1",
				},
			},
		}),
	)
	return container, err
}

func CreateDynamoDBClient(t *testing.T, ctx context.Context, l *localstack.LocalStackContainer) (*dynamodb.Client, error) {
	endpoint, err := l.PortEndpoint(ctx, "4566/tcp", "http")
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("x", "x", "x")),
	)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(awsCfg, func(options *dynamodb.Options) {
		options.BaseEndpoint = aws.String(endpoint)
	})
	t.Log("created dynamodb client")
	return client, nil
}

// CreateJournalTable creates the v5 journal table. The journal has no Streams:
// the change feed comes from the head table (DY-3).
func CreateJournalTable(t *testing.T, ctx context.Context, client *dynamodb.Client, tableName string) error {
	t.Helper()
	return createTableEventually(t, func() error {
		_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(tableName),
			AttributeDefinitions: []types.AttributeDefinition{
				{AttributeName: aws.String("aid"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("seq_nr"), AttributeType: types.ScalarAttributeTypeN},
			},
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("aid"), KeyType: types.KeyTypeHash},
				{AttributeName: aws.String("seq_nr"), KeyType: types.KeyTypeRange},
			},
			BillingMode: types.BillingModePayPerRequest,
		})
		return err
	})
}

func CreateSnapshotTable(t *testing.T, ctx context.Context, client *dynamodb.Client, tableName string, historyIndexName string) error {
	t.Helper()
	return createTableEventually(t, func() error {
		_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(tableName),
			AttributeDefinitions: []types.AttributeDefinition{
				{AttributeName: aws.String("aid"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("skey"), AttributeType: types.ScalarAttributeTypeN},
				{AttributeName: aws.String("active_history_seq_nr"), AttributeType: types.ScalarAttributeTypeN},
			},
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("aid"), KeyType: types.KeyTypeHash},
				{AttributeName: aws.String("skey"), KeyType: types.KeyTypeRange},
			},
			BillingMode: types.BillingModePayPerRequest,
			GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String(historyIndexName),
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("aid"), KeyType: types.KeyTypeHash},
						{AttributeName: aws.String("active_history_seq_nr"), KeyType: types.KeyTypeRange},
					},
					Projection: &types.Projection{ProjectionType: types.ProjectionTypeKeysOnly},
				},
			},
		})
		return err
	})
}

func CreateHeadTable(t *testing.T, ctx context.Context, client *dynamodb.Client, tableName string) error {
	t.Helper()
	return createTableEventually(t, func() error {
		_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(tableName),
			AttributeDefinitions: []types.AttributeDefinition{
				{AttributeName: aws.String("aid"), AttributeType: types.ScalarAttributeTypeS},
			},
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("aid"), KeyType: types.KeyTypeHash},
			},
			BillingMode: types.BillingModePayPerRequest,
			StreamSpecification: &types.StreamSpecification{
				StreamEnabled:  aws.Bool(true),
				StreamViewType: types.StreamViewTypeNewImage,
			},
		})
		return err
	})
}

// createTableEventually retries CreateTable because LocalStack can reset the
// connection while its DynamoDB provider is still warming up. ResourceInUse
// means another attempt already created the table.
func createTableEventually(t *testing.T, create func() error) error {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for {
		err := create()
		if err == nil {
			return nil
		}
		var inUse *types.ResourceInUseException
		if errors.As(err, &inUse) {
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(time.Second)
	}
}

// WaitForDynamoDB waits until LocalStack answers DynamoDB requests. The health
// endpoint can report ready before the DynamoDB provider accepts calls.
func WaitForDynamoDB(t *testing.T, ctx context.Context, client *dynamodb.Client) {
	t.Helper()
	require.Eventually(t, func() bool {
		_, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
		return err == nil
	}, 60*time.Second, time.Second)
}

func GetDataSourceName(port network.Port) string {
	dbUrl := fmt.Sprintf("ceer:ceer@tcp(localhost:%s)/ceer", port.Port())
	dataSourceName := fmt.Sprintf("%s?parseTime=true", dbUrl)
	return dataSourceName
}

func MigrateDB(t *testing.T, err error, db *sqlx.DB, position string) error {
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	require.NoError(t, err)
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/tools/migrate/migrations", position),
		"mysql",
		driver,
	)
	require.NoError(t, err)
	err = m.Steps(5)
	require.NoError(t, err)
	return err
}
