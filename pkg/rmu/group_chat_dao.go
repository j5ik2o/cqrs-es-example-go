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

// NewGroupChatDaoImpl は GroupChatDaoImpl を生成します。
func NewGroupChatDaoImpl(db *sqlx.DB) *GroupChatDaoImpl {
	return &GroupChatDaoImpl{db}
}

// Create は DB上にグループチャットリードモデルを作成します。
func (dao *GroupChatDaoImpl) Create(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO group_chats (id, name, owner_id, created_at) VALUES(?,?,?,?)`)
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}(stmt)
	if err != nil {
		return err
	}
	dt := createdAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(aggregateId.AsString(), name.String(), administratorId.AsString(), dt)
	if err != nil {
		return err
	}
	return nil
}

// AddMember は DB上にメンバーリードモデルを追加します。
func (dao *GroupChatDaoImpl) AddMember(id *models.MemberId, aggregateId *models.GroupChatId, accountId *models.UserAccountId, role models.Role, at time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO members (id, group_chat_id, account_id, role, created_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(id.String(), aggregateId.AsString(), accountId.AsString(), role.String(), dt)
	if err != nil {
		return err
	}
	return nil
}

func (dao *GroupChatDaoImpl) InsertMessage(id *models.MessageId, groupChatId *models.GroupChatId, accountId *models.UserAccountId, text string, at time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO messages (id, group_chat_id, account_id, text, created_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	dt := at.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(id.String(), groupChatId.AsString(), accountId.AsString(), text, dt)
	if err != nil {
		return err
	}
	return nil
}
