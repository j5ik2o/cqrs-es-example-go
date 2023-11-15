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
	tx := dao.db.MustBegin()
	tx.MustExec("INSERT INTO group_chats (id, name, owner_id, created_at) VALUES($1,$2,$3,$4)", aggregateId.AsString(), name.String(), administratorId.AsString(), createdAt)
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
