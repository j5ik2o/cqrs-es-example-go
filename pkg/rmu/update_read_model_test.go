package rmu

import (
	"context"
	"cqrs-es-example-go/test"
	_ "embed"
	"encoding/json"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:embed example-dynamodb-event.json
var eventData []byte

func TestUpdateReadModel(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

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

	err = test.MigrateDB(t, err, db, "../../")

	dao := NewGroupChatDaoImpl(db)
	var parsed dynamodbevents.DynamoDBEvent
	err = json.Unmarshal(eventData, &parsed)
	require.NoError(t, err)
	readModelUpdater := NewReadModelUpdater(dao)
	err = readModelUpdater.UpdateReadModel(context.Background(), parsed)
	require.NoError(t, err)

}
