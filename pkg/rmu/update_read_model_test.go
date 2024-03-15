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
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)
	require.NoError(t, err)

	err = test.MigrateDB(t, err, db, "../../")

	dao := NewGroupChatDaoImpl(db)
	var parsed dynamodbevents.DynamoDBEvent
	err = json.Unmarshal(eventData, &parsed)
	require.NoError(t, err)
	readModelUpdater := NewReadModelUpdater(&dao)
	err = readModelUpdater.UpdateReadModel(context.Background(), parsed)
	require.NoError(t, err)

}
