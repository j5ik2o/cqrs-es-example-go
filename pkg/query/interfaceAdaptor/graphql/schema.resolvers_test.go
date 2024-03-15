package querygraphql

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
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	member := models.NewMember(memberId, adminId, models.AdminRole)
	err = dao.InsertMember(&groupChatId, &member, now)

	result, err := resolver.QueryRoot().GetGroupChat(ctx, groupChatId.AsString(), adminId.AsString())
	require.NoError(t, err)
	assert.Equal(t, groupChatId.AsString(), result.ID)
}

func Test_GetGroupChats(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)

	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	member := models.NewMember(memberId, adminId, models.AdminRole)
	err = dao.InsertMember(&groupChatId, &member, now)

	result, err := resolver.QueryRoot().GetGroupChats(ctx, adminId.AsString())
	require.NoError(t, err)
	assert.Equal(t, groupChatId.AsString(), result[0].ID)
}

func Test_GetMember(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	member := models.NewMember(memberId, adminId, models.AdminRole)
	err = dao.InsertMember(&groupChatId, &member, now)

	result, err := resolver.QueryRoot().GetMember(ctx, groupChatId.AsString(), adminId.AsString())
	require.NoError(t, err)
	require.Equal(t, groupChatId.AsString(), result.GroupChatID)
	require.Equal(t, adminId.AsString(), result.UserAccountID)
}

func Test_GetMembers(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	member := models.NewMember(memberId, adminId, models.AdminRole)
	err = dao.InsertMember(&groupChatId, &member, now)

	results, err := resolver.QueryRoot().GetMembers(ctx, groupChatId.AsString(), adminId.AsString())
	require.NoError(t, err)
	require.Equal(t, groupChatId.AsString(), results[0].GroupChatID)
	require.Equal(t, adminId.AsString(), results[0].UserAccountID)
}

func Test_GetMessage(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	member := models.NewMember(memberId, adminId, models.AdminRole)
	err = dao.InsertMember(&groupChatId, &member, now)
	messageId := models.NewMessageId()
	err = dao.InsertMessage(&messageId, &groupChatId, &adminId, "test", now)

	result, err := resolver.QueryRoot().GetMessage(ctx, messageId.String(), adminId.AsString())
	require.NoError(t, err)
	require.Equal(t, groupChatId.AsString(), result.GroupChatID)
	require.Equal(t, adminId.AsString(), result.UserAccountID)
	require.Equal(t, messageId.String(), result.ID)
	require.Equal(t, "test", result.Text)
}

func Test_GetMessages(t *testing.T) {
	ctx := context.Background()
	container, err := test.CreateMySQLContainer(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			require.NoError(t, err)
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../../../")

	var resolver ResolverRoot = NewResolver(db)

	dao := rmu.NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()

	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)
	memberId := models.NewMemberId()
	member := models.NewMember(memberId, adminId, models.AdminRole)
	err = dao.InsertMember(&groupChatId, &member, now)
	messageId := models.NewMessageId()
	err = dao.InsertMessage(&messageId, &groupChatId, &adminId, "test", now)

	results, err := resolver.QueryRoot().GetMessages(ctx, groupChatId.AsString(), adminId.AsString())
	require.NoError(t, err)
	require.Equal(t, groupChatId.AsString(), results[0].GroupChatID)
	require.Equal(t, adminId.AsString(), results[0].UserAccountID)
	require.Equal(t, messageId.String(), results[0].ID)
	require.Equal(t, "test", results[0].Text)
}
