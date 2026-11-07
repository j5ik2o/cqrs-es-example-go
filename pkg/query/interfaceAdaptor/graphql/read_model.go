package querygraphql

import (
	"context"
	"time"

	querygraphql "cqrs-es-example-go/pkg/query/interfaceAdaptor/graphql/model"
)

// naiveDateTimeLayout is the ISO 8601 layout without a timezone that the
// NaiveDateTime scalar advertises in schema.graphqls.
const naiveDateTimeLayout = "2006-01-02T15:04:05.999999999"

func formatNaiveDateTime(value time.Time) string {
	return value.Format(naiveDateTimeLayout)
}

type groupChatRow struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	OwnerID   string    `db:"owner_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (row groupChatRow) toGraphQL() *querygraphql.GroupChat {
	return &querygraphql.GroupChat{
		ID:        row.ID,
		Name:      row.Name,
		OwnerID:   row.OwnerID,
		CreatedAt: formatNaiveDateTime(row.CreatedAt),
		UpdatedAt: formatNaiveDateTime(row.UpdatedAt),
	}
}

type memberRow struct {
	ID            string    `db:"id"`
	GroupChatID   string    `db:"group_chat_id"`
	UserAccountID string    `db:"user_account_id"`
	Role          string    `db:"role"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (row memberRow) toGraphQL() *querygraphql.Member {
	return &querygraphql.Member{
		ID:            row.ID,
		GroupChatID:   row.GroupChatID,
		UserAccountID: row.UserAccountID,
		Role:          row.Role,
		CreatedAt:     formatNaiveDateTime(row.CreatedAt),
		UpdatedAt:     formatNaiveDateTime(row.UpdatedAt),
	}
}

type messageRow struct {
	ID            string    `db:"id"`
	GroupChatID   string    `db:"group_chat_id"`
	UserAccountID string    `db:"user_account_id"`
	Text          string    `db:"text"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (row messageRow) toGraphQL() *querygraphql.Message {
	return &querygraphql.Message{
		ID:            row.ID,
		GroupChatID:   row.GroupChatID,
		UserAccountID: row.UserAccountID,
		Text:          row.Text,
		CreatedAt:     formatNaiveDateTime(row.CreatedAt),
		UpdatedAt:     formatNaiveDateTime(row.UpdatedAt),
	}
}

func (r *Resolver) findGroupChat(ctx context.Context, groupChatID, userAccountID string) (*groupChatRow, error) {
	const query = `SELECT gc.id, gc.name, gc.owner_id, gc.created_at, gc.updated_at
		FROM group_chats AS gc
		JOIN members AS m ON gc.id = m.group_chat_id
		WHERE gc.disabled = 0 AND m.group_chat_id = ? AND m.user_account_id = ?`
	var row groupChatRow
	if err := r.db.GetContext(ctx, &row, query, groupChatID, userAccountID); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Resolver) findGroupChats(ctx context.Context, userAccountID string) ([]groupChatRow, error) {
	const query = `SELECT gc.id, gc.name, gc.owner_id, gc.created_at, gc.updated_at
		FROM group_chats AS gc
		JOIN members AS m ON gc.id = m.group_chat_id
		WHERE gc.disabled = 0 AND m.user_account_id = ?
		ORDER BY gc.created_at, gc.id`
	var rows []groupChatRow
	if err := r.db.SelectContext(ctx, &rows, query, userAccountID); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Resolver) findMember(ctx context.Context, groupChatID, userAccountID string) (*memberRow, error) {
	const query = `SELECT m.id, m.group_chat_id, m.user_account_id, m.role, m.created_at, m.updated_at
		FROM group_chats AS gc
		JOIN members AS m ON gc.id = m.group_chat_id
		WHERE gc.disabled = 0 AND m.group_chat_id = ? AND m.user_account_id = ?`
	var row memberRow
	if err := r.db.GetContext(ctx, &row, query, groupChatID, userAccountID); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Resolver) findMembers(ctx context.Context, groupChatID, userAccountID string) ([]memberRow, error) {
	const query = `SELECT m.id, m.group_chat_id, m.user_account_id, m.role, m.created_at, m.updated_at
		FROM group_chats AS gc
		JOIN members AS m ON gc.id = m.group_chat_id
		WHERE gc.disabled = 0 AND m.group_chat_id = ?
			AND EXISTS (SELECT 1 FROM members AS requester
				WHERE requester.group_chat_id = m.group_chat_id AND requester.user_account_id = ?)
		ORDER BY m.created_at, m.id`
	var rows []memberRow
	if err := r.db.SelectContext(ctx, &rows, query, groupChatID, userAccountID); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Resolver) findMessage(ctx context.Context, messageID, userAccountID string) (*messageRow, error) {
	const query = `SELECT m.id, m.group_chat_id, m.user_account_id, m.text, m.created_at, m.updated_at
		FROM group_chats AS gc
		JOIN messages AS m ON gc.id = m.group_chat_id
		WHERE gc.disabled = 0 AND m.disabled = 0 AND m.id = ?
			AND EXISTS (SELECT 1 FROM members AS requester
				WHERE requester.group_chat_id = m.group_chat_id AND requester.user_account_id = ?)`
	var row messageRow
	if err := r.db.GetContext(ctx, &row, query, messageID, userAccountID); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Resolver) findMessages(ctx context.Context, groupChatID, userAccountID string) ([]messageRow, error) {
	const query = `SELECT m.id, m.group_chat_id, m.user_account_id, m.text, m.created_at, m.updated_at
		FROM group_chats AS gc
		JOIN messages AS m ON gc.id = m.group_chat_id
		WHERE gc.disabled = 0 AND m.disabled = 0 AND m.group_chat_id = ?
			AND EXISTS (SELECT 1 FROM members AS requester
				WHERE requester.group_chat_id = m.group_chat_id AND requester.user_account_id = ?)
		ORDER BY m.created_at, m.id`
	var rows []messageRow
	if err := r.db.SelectContext(ctx, &rows, query, groupChatID, userAccountID); err != nil {
		return nil, err
	}
	return rows, nil
}
