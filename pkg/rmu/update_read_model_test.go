package rmu

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	dynamodbevents "github.com/aws/aws-lambda-go/events"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/env"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

//go:embed example-dynamodb-event.json
var eventData []byte

func TestUpdateReadModel(t *testing.T) {
	err := os.Setenv("DATABASE_URL", "ceer:ceer@tcp(localhost:3306)/ceer")
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
