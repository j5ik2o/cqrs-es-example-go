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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainermysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"testing"
)

func StartContainer(t *testing.T, ctx context.Context) (error, nat.Port) {
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
	return err, port
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
