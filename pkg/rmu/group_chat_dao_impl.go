package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
)

// sqlPreparer is the shared subset of *sqlx.DB and *sqlx.Tx used by the write
// methods, so the same SQL runs inside and outside a transaction.
type sqlPreparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

type GroupChatDaoImpl struct {
	db *sqlx.DB
}

func NewGroupChatDaoImpl(db *sqlx.DB) *GroupChatDaoImpl {
	return &GroupChatDaoImpl{db}
}

func (dao *GroupChatDaoImpl) GetPosition(ctx context.Context, aggregateID string) (int64, error) {
	var seqNr int64
	err := dao.db.GetContext(ctx, &seqNr, `SELECT seq_nr FROM read_model_positions WHERE aggregate_id = ?`, aggregateID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return seqNr, nil
}

func (dao *GroupChatDaoImpl) Process(ctx context.Context, aggregateID string, fn func(w ReadModelWriter, currentSeqNr int64) (int64, error)) error {
	tx, err := dao.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var currentSeqNr int64
	err = tx.GetContext(ctx, &currentSeqNr, `SELECT seq_nr FROM read_model_positions WHERE aggregate_id = ? FOR UPDATE`, aggregateID)
	if errors.Is(err, sql.ErrNoRows) {
		currentSeqNr = 0
	} else if err != nil {
		return err
	}

	writer := &readModelTx{
		groupChatWriter: &groupChatWriter{preparer: tx},
		tx:              tx,
	}
	nextSeqNr, err := fn(writer, currentSeqNr)
	if err != nil {
		return err
	}
	if nextSeqNr != currentSeqNr {
		if err := writer.updatePosition(ctx, aggregateID, nextSeqNr); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// readModelTx adds the position write to the writer scoped to a transaction.
type readModelTx struct {
	*groupChatWriter
	tx *sqlx.Tx
}

func (w *readModelTx) updatePosition(ctx context.Context, aggregateID string, seqNr int64) error {
	_, err := w.tx.ExecContext(ctx,
		`INSERT INTO read_model_positions (aggregate_id, seq_nr, updated_at) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE seq_nr = VALUES(seq_nr), updated_at = VALUES(updated_at)`,
		aggregateID, seqNr, time.Now())
	return err
}

// groupChatWriter holds the read model SQL. GroupChatDaoImpl delegates its
// direct methods here so the tests can seed the read model without a transaction.
type groupChatWriter struct {
	preparer sqlPreparer
}

// Direct writes on the DAO delegate to the same SQL used inside Process.
func (dao *GroupChatDaoImpl) InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).InsertGroupChat(aggregateId, name, administratorId, createdAt)
}

func (dao *GroupChatDaoImpl) DeleteGroupChat(aggregateId *models.GroupChatId, updatedAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).DeleteGroupChat(aggregateId, updatedAt)
}

func (dao *GroupChatDaoImpl) UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, updatedAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).UpdateName(aggregateId, name, updatedAt)
}

func (dao *GroupChatDaoImpl) InsertMember(aggregateId *models.GroupChatId, member *models.Member, createdAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).InsertMember(aggregateId, member, createdAt)
}

func (dao *GroupChatDaoImpl) DeleteMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId) error {
	return (&groupChatWriter{preparer: dao.db}).DeleteMember(groupChatId, userAccountId)
}

func (dao *GroupChatDaoImpl) InsertMessage(messageId *models.MessageId, aggregateId *models.GroupChatId, userAccountId *models.UserAccountId, text string, createdAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).InsertMessage(messageId, aggregateId, userAccountId, text, createdAt)
}

func (dao *GroupChatDaoImpl) UpdateMessage(messageId *models.MessageId, text string, updatedAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).UpdateMessage(messageId, text, updatedAt)
}

func (dao *GroupChatDaoImpl) DeleteMessage(messageId *models.MessageId, updatedAt time.Time) error {
	return (&groupChatWriter{preparer: dao.db}).DeleteMessage(messageId, updatedAt)
}

func (w *groupChatWriter) InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error {
	stmt, err := w.preparer.Prepare(`INSERT INTO group_chats (id, disabled, name, owner_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := createdAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(aggregateId.AsString(), false, name.String(), administratorId.AsString(), dt, dt)
	return err
}

func (w *groupChatWriter) DeleteGroupChat(aggregateId *models.GroupChatId, updatedAt time.Time) error {
	stmt, err := w.preparer.Prepare(`UPDATE group_chats SET disabled = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := updatedAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(true, dt, aggregateId.AsString())
	return err
}

func (w *groupChatWriter) UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, updatedAt time.Time) error {
	stmt, err := w.preparer.Prepare(`UPDATE group_chats SET name = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := updatedAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(name.String(), dt, aggregateId.AsString())
	return err
}

func (w *groupChatWriter) InsertMember(aggregateId *models.GroupChatId, member *models.Member, createdAt time.Time) error {
	stmt, err := w.preparer.Prepare(`INSERT INTO members (id, group_chat_id, user_account_id, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := createdAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(member.GetId().String(), aggregateId.AsString(), member.GetUserAccountId().AsString(), member.GetRole().String(), dt, dt)
	return err
}

func (w *groupChatWriter) DeleteMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId) error {
	stmt, err := w.preparer.Prepare(`DELETE FROM members WHERE group_chat_id = ? AND user_account_id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	_, err = stmt.Exec(groupChatId.AsString(), userAccountId.AsString())
	return err
}

func (w *groupChatWriter) InsertMessage(messageId *models.MessageId, groupChatId *models.GroupChatId, userAccountId *models.UserAccountId, text string, createdAt time.Time) error {
	stmt, err := w.preparer.Prepare(`INSERT INTO messages (id, disabled, group_chat_id, user_account_id, text, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := createdAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(messageId.String(), false, groupChatId.AsString(), userAccountId.AsString(), text, dt, dt)
	return err
}

func (w *groupChatWriter) UpdateMessage(messageId *models.MessageId, text string, updatedAt time.Time) error {
	stmt, err := w.preparer.Prepare(`UPDATE messages SET text = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := updatedAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(text, dt, messageId.String())
	return err
}

func (w *groupChatWriter) DeleteMessage(messageId *models.MessageId, updatedAt time.Time) error {
	stmt, err := w.preparer.Prepare(`UPDATE messages SET disabled = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	dt := updatedAt.Format("2006-01-02 15:04:05")
	_, err = stmt.Exec(true, dt, messageId.String())
	return err
}
