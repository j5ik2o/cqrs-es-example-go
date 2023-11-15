package rmu

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type GroupChatDao struct {
	db *sqlx.DB
}

func NewGroupChatDao(db *sqlx.DB) *GroupChatDao {
	return &GroupChatDao{db}
}

func (dao *GroupChatDao) Create(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO group_chats (id, name, owner_id, created_at) VALUES(?,?,?,?)`)
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
