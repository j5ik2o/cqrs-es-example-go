package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"cqrs-es-example-go/test"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestReadModelDao_InsertGroupChat(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	groupChat, err := getGroupChat(db, groupChatId)
	require.NoError(t, err)
	require.NotNil(t, groupChat)
	require.Equal(t, groupChatId.AsString(), groupChat["ID"])
	require.Equal(t, false, groupChat["Disabled"])
	require.Equal(t, groupChatName.String(), groupChat["Name"])
	require.Equal(t, adminId.AsString(), groupChat["OwnerID"])
}

func TestReadModelDao_DeleteGroupChat(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	groupChat, err := getGroupChat(db, groupChatId)
	require.NoError(t, err)
	require.NotNil(t, groupChat)
	require.Equal(t, groupChatId.AsString(), groupChat["ID"])
	require.Equal(t, false, groupChat["Disabled"])
	require.Equal(t, groupChatName.String(), groupChat["Name"])
	require.Equal(t, adminId.AsString(), groupChat["OwnerID"])

	err = dao.DeleteGroupChat(&groupChatId, now)
	require.NoError(t, err)

	groupChat, err = getGroupChat(db, groupChatId)
	require.NoError(t, err)
	require.NotNil(t, groupChat)
	require.Equal(t, groupChatId.AsString(), groupChat["ID"])
	require.Equal(t, true, groupChat["Disabled"])
	require.Equal(t, groupChatName.String(), groupChat["Name"])
	require.Equal(t, adminId.AsString(), groupChat["OwnerID"])
}

func TestReadModelDao_RenameGroupChat(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	groupChatName = models.NewGroupChatName("test2").MustGet()
	err = dao.UpdateName(&groupChatId, &groupChatName, now)
	require.NoError(t, err)

	groupChat, err := getGroupChat(db, groupChatId)
	require.NoError(t, err)
	require.NotNil(t, groupChat)
	require.Equal(t, groupChatId.AsString(), groupChat["ID"])
	require.Equal(t, groupChatName.String(), groupChat["Name"])
	require.Equal(t, adminId.AsString(), groupChat["OwnerID"])
}

func TestReadModelDao_InsertMember(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	memberId := models.NewMemberId()
	err = dao.InsertMember(&memberId, &groupChatId, &adminId, models.AdminRole, now)
	require.NoError(t, err)

	member, err := getMember(db, groupChatId, adminId)
	require.NoError(t, err)
	require.NotNil(t, member)
	require.Equal(t, memberId.String(), member["ID"])
}

func TestReadModelDao_RemoveMember(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	memberId := models.NewMemberId()
	err = dao.InsertMember(&memberId, &groupChatId, &adminId, models.AdminRole, now)
	require.NoError(t, err)

	member, err := getMember(db, groupChatId, adminId)
	require.NoError(t, err)
	require.NotNil(t, member)
	require.Equal(t, memberId.String(), member["ID"])

	err = dao.DeleteMember(&groupChatId, &adminId)
	require.NoError(t, err)

	member, err = getMember(db, groupChatId, adminId)
	require.Error(t, err)
}

func TestReadModelDao_InsertMessage(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	memberId := models.NewMemberId()
	err = dao.InsertMember(&memberId, &groupChatId, &adminId, models.AdminRole, now)
	require.NoError(t, err)

	messageId := models.NewMessageId()
	text := "test"
	err = dao.InsertMessage(&messageId, &groupChatId, &adminId, text, now)
	require.NoError(t, err)

	message, err := getMessage(db, messageId)
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Equal(t, messageId.String(), message["ID"])
	require.Equal(t, false, message["Disabled"])
}

func TestReadModelDao_DeleteMessage(t *testing.T) {
	ctx := context.Background()
	err, port := test.StartContainer(t, ctx)
	dataSourceName := test.GetDataSourceName(port)

	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	require.NoError(t, err)
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}(db)

	err = test.MigrateDB(t, err, db, "../../")
	require.NoError(t, err)

	dao := NewGroupChatDaoImpl(db)
	groupChatId := models.NewGroupChatId()
	groupChatName := models.NewGroupChatName("test").MustGet()
	adminId := models.NewUserAccountId()
	now := time.Now()
	err = dao.InsertGroupChat(&groupChatId, &groupChatName, &adminId, now)
	require.NoError(t, err)

	memberId := models.NewMemberId()
	err = dao.InsertMember(&memberId, &groupChatId, &adminId, models.AdminRole, now)
	require.NoError(t, err)

	messageId := models.NewMessageId()
	text := "test"
	err = dao.InsertMessage(&messageId, &groupChatId, &adminId, text, now)
	require.NoError(t, err)

	err = dao.DeleteMessage(&messageId, now)
	require.NoError(t, err)

	message, err := getMessage(db, messageId)
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Equal(t, messageId.String(), message["ID"])
	require.Equal(t, true, message["Disabled"])
}

func getGroupChat(db *sqlx.DB, groupChatId models.GroupChatId) (map[string]any, error) {
	stmt, err := db.Prepare(`SELECT gc.id, gc.disabled, gc.name, gc.owner_id, gc.created_at, gc.updated_at FROM group_chats AS gc WHERE gc.id = ?`)
	if err != nil {
		return nil, err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	row := stmt.QueryRow(groupChatId.AsString())
	if row != nil {
		var id string
		var disabled bool
		var name string
		var ownerID string
		var createdAt time.Time
		var updatedAt time.Time
		err = row.Scan(&id, &disabled, &name, &ownerID, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"ID":        id,
			"Disabled":  disabled,
			"Name":      name,
			"OwnerID":   ownerID,
			"CreatedAt": createdAt.String(),
			"UpdatedAt": updatedAt.String(),
		}, nil
	}
	return nil, nil
}

func getMember(db *sqlx.DB, groupChatId models.GroupChatId, accountId models.UserAccountId) (map[string]any, error) {
	stmt, err := db.Prepare(`SELECT m.id, m.role, m.created_at, m.updated_at FROM members AS m WHERE m.group_chat_id = ? AND m.user_account_id = ?`)
	if err != nil {
		return nil, err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	row := stmt.QueryRow(groupChatId.AsString(), accountId.AsString())
	if row != nil {
		var id string
		var role string
		var createdAt time.Time
		var updatedAt time.Time
		err = row.Scan(&id, &role, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"ID":        id,
			"Role":      role,
			"CreatedAt": createdAt.String(),
			"UpdatedAt": updatedAt.String(),
		}, nil
	}
	return nil, nil
}

func getMessage(db *sqlx.DB, messageId models.MessageId) (map[string]any, error) {
	stmt, err := db.Prepare(`SELECT m.id, m.disabled, m.group_chat_id, m.user_account_id, m.text, m.created_at, m.updated_at FROM messages AS m WHERE m.id = ?`)
	if err != nil {
		return nil, err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	row := stmt.QueryRow(messageId.String())
	if row != nil {
		var id string
		var groupChatId string
		var disabled bool
		var userAccountId string
		var text string
		var createdAt time.Time
		var updatedAt time.Time
		err = row.Scan(&id, &disabled, &groupChatId, &userAccountId, &text, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"ID":            id,
			"Disabled":      disabled,
			"GroupChatID":   groupChatId,
			"UserAccountID": userAccountId,
			"Text":          text,
			"CreatedAt":     createdAt.String(),
			"UpdatedAt":     updatedAt.String(),
		}, nil
	}
	return nil, nil
}
