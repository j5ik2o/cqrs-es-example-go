package graph

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/pkg/rmu"
	"cqrs-es-example-go/test"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_GetGroupChat(t *testing.T) {
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

	err = test.MigrateDB(t, err, db, "../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	err = dao.InsertMember(&memberId, &groupChatId, &adminId, models.AdminRole, now)

	result, err := resolver.QueryRoot().GetGroupChat(ctx, groupChatId.AsString(), adminId.AsString())
	require.NoError(t, err)
	assert.Equal(t, groupChatId.AsString(), result.ID)
}

func Test_GetGroupChats(t *testing.T) {
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

	err = test.MigrateDB(t, err, db, "../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	err = dao.InsertMember(&memberId, &groupChatId, &adminId, models.AdminRole, now)

	result, err := resolver.QueryRoot().GetGroupChats(ctx, adminId.AsString())
	require.NoError(t, err)
	assert.Equal(t, groupChatId.AsString(), result[0].ID)
}
