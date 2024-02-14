package rmu

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

type GroupChatDaoImpl struct {
	db *sqlx.DB
}

func NewGroupChatDaoImpl(db *sqlx.DB) GroupChatDaoImpl {
	return GroupChatDaoImpl{db}
}

func (dao *GroupChatDaoImpl) InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO group_chats (id, disabled, name, owner_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	dt := createdAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(aggregateId.AsString(), false, name.String(), administratorId.AsString(), dt, dt)
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) DeleteGroupChat(aggregateId *models.GroupChatId, at time.Time) error {
	stmt, err := dao.db.Prepare(`UPDATE group_chats SET disabled = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(true, dt, aggregateId.AsString())
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, at time.Time) error {
	stmt, err := dao.db.Prepare(`UPDATE group_chats SET name = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(name.String(), dt, aggregateId.AsString())
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) InsertMember(id *models.MemberId, aggregateId *models.GroupChatId, userAccountId *models.UserAccountId, role models.Role, at time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO members (id, group_chat_id, user_account_id, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(id.String(), aggregateId.AsString(), userAccountId.AsString(), role.String(), dt, dt)
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) DeleteMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId) error {
	stmt, err := dao.db.Prepare(`DELETE FROM members WHERE group_chat_id = ? AND user_account_id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	_, err = stmt.Exec(groupChatId.AsString(), userAccountId.AsString())
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) InsertMessage(id *models.MessageId, groupChatId *models.GroupChatId, userAccountId *models.UserAccountId, text string, at time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO messages (id, disabled, group_chat_id, user_account_id, text, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(id.String(), false, groupChatId.AsString(), userAccountId.AsString(), text, dt, dt)
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) DeleteMessage(id *models.MessageId, at time.Time) error {
	stmt, err := dao.db.Prepare(`UPDATE messages SET disabled = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(true, dt, id.String())
	if err != nil {
		return err
	}
	return nil
}
