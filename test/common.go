package test

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	testcontainermysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"testing"
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
					"SERVICES":              "dynamodb",
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

func GetDataSourceName(port nat.Port) string {
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
	err = m.Steps(3)
	require.NoError(t, err)
	return err
}
