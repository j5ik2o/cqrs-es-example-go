package rmu

import (
	"cqrs-es-example-go/pkg/command/domain/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type GroupChatDaoImpl struct {
	db *sqlx.DB
}

func NewGroupChatDaoImpl(db *sqlx.DB) *GroupChatDaoImpl {
	return &GroupChatDaoImpl{db}
}

func (dao *GroupChatDaoImpl) Create(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error {
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

func (dao *GroupChatDaoImpl) AddMember(id *models.MemberId, aggregateId *models.GroupChatId, accountId *models.UserAccountId, role models.Role, at time.Time) error {
	stmt, err := dao.db.Prepare(`INSERT INTO members (id, group_chat_id, account_id, role, created_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	dt := at.Format("2006-01-02 15:04:05")
	var roleStr string
	switch role {
	case models.AdminRole:
		roleStr = "admin"
	case models.MemberRole:
		roleStr = "member"
	}
	_, err = stmt.Exec(id.String(), aggregateId.AsString(), accountId.AsString(), roleStr, dt)
	if err != nil {
		return err
	}
	return nil
}
