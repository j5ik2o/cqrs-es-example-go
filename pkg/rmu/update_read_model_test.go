package rmu

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainermysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:embed example-dynamodb-event.json
var eventData []byte

func TestUpdateReadModel(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainermysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.2"),
		// mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		testcontainermysql.WithDatabase("ceer"),
		testcontainermysql.WithUsername("ceer"),
		testcontainermysql.WithPassword("ceer"),
		// testcontainermysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	require.NoError(t, err)
	assert.NotNil(t, container)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)

	dbUrl := fmt.Sprintf("ceer:ceer@tcp(localhost:%s)/ceer", port.Port())
	dataSourceName := fmt.Sprintf("%s?parseTime=true", dbUrl)

	db, err := sqlx.Connect("mysql", dataSourceName)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)

	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	require.NoError(t, err)
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../tools/migrate/migrations",
		"mysql",
		driver,
	)
	require.NoError(t, err)
	err = m.Steps(3)
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	var parsed dynamodbevents.DynamoDBEvent
	err = json.Unmarshal(eventData, &parsed)
	require.NoError(t, err)
	readModelUpdater := NewReadModelUpdater(dao)
	err = readModelUpdater.UpdateReadModel(context.Background(), parsed)
	require.NoError(t, err)

}
