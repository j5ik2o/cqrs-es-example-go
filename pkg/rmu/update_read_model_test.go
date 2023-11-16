package rmu

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"os"
	"path/filepath"
	"testing"
)

//go:embed example-dynamodb-event.json
var eventData []byte

func TestUpdateReadModel(t *testing.T) {
	ctx := context.Background()
	container, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8"),
		// mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		mysql.WithDatabase("ceer"),
		mysql.WithUsername("ceer"),
		mysql.WithPassword("ceer"),
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	require.NoError(t, err)
	assert.NotNil(t, container)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	err = os.Setenv("DATABASE_URL", fmt.Sprintf("ceer:ceer@tcp(localhost:%s)/ceer", port.Port()))
	require.NoError(t, err)
	dbUrl := env.String("", "DATABASE_URL")
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
	dao := NewGroupChatDaoImpl(db)

	var parsed dynamodbevents.DynamoDBEvent
	err = json.Unmarshal(eventData, &parsed)
	require.NoError(t, err)
	readModelUpdater := NewReadModelUpdater(dao)
	err = readModelUpdater.UpdateReadModel(context.Background(), parsed)
	require.NoError(t, err)

}
