package rmu

import (
	"context"
	"cqrs-es-example-go/pkg/command/domain/models"
	"time"
)

// GroupChatDao applies one aggregate's events to the read model. It records the
// processed position together with the read model change so that a restart or a
// retried batch resumes at the last committed event.
type GroupChatDao interface {
	// GetPosition returns the last applied seq_nr for the aggregate, or 0 when
	// no event has been applied yet.
	GetPosition(ctx context.Context, aggregateID string) (int64, error)

	// Process runs fn inside one database transaction. It locks the aggregate's
	// position row, passes the current position to fn, and writes the position
	// fn returns in the same transaction. The read model changes fn makes
	// through the writer commit or roll back together with that position.
	Process(ctx context.Context, aggregateID string, fn func(w ReadModelWriter, currentSeqNr int64) (int64, error)) error
}

// ReadModelWriter is the read model write surface inside the transaction opened
// by GroupChatDao.Process.
type ReadModelWriter interface {
	InsertGroupChat(aggregateId *models.GroupChatId, name *models.GroupChatName, administratorId *models.UserAccountId, createdAt time.Time) error
	DeleteGroupChat(aggregateId *models.GroupChatId, updatedAt time.Time) error
	UpdateName(aggregateId *models.GroupChatId, name *models.GroupChatName, updatedAt time.Time) error

	InsertMember(aggregateId *models.GroupChatId, member *models.Member, createdAt time.Time) error
	DeleteMember(groupChatId *models.GroupChatId, userAccountId *models.UserAccountId) error

	InsertMessage(messageId *models.MessageId, aggregateId *models.GroupChatId, userAccountId *models.UserAccountId, text string, createdAt time.Time) error
	UpdateMessage(messageId *models.MessageId, text string, updatedAt time.Time) error
	DeleteMessage(messageId *models.MessageId, updatedAt time.Time) error
}
