package graph

import (
	"context"
	query "cqrs-es-example-go/pkg/query/graph/model"
	"fmt"
	"time"
)

// GetGroupChat is the resolver for the getGroupChat field.
func (r *queryRootResolver) GetGroupChat(ctx context.Context, groupChatID string, accountID string) (*query.GroupChat, error) {
	stmt, err := r.db.Prepare(`SELECT gc.id, gc.name, gc.owner_id, gc.created_at FROM group_chats AS gc JOIN members AS m ON gc.id = m.group_chat_id WHERE m.group_chat_id = ? AND m.account_id = ?`)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(groupChatID, accountID)
	if row != nil {
		var id string
		var name string
		var ownerID string
		var createdAt time.Time
		err = row.Scan(&id, &name, &ownerID, &createdAt)
		if err != nil {
			return nil, err
		}
		return &query.GroupChat{
			ID:        id,
			Name:      name,
			OwnerID:   ownerID,
			CreatedAt: createdAt.String(),
		}, nil
	}
	return nil, nil
}

// GetGroupChats is the resolver for the getGroupChats field.
func (r *queryRootResolver) GetGroupChats(ctx context.Context, accountID string) ([]*query.GroupChat, error) {
	panic(fmt.Errorf("not implemented: GetGroupChats - getGroupChats"))
}

// GetMember is the resolver for the getMember field.
func (r *queryRootResolver) GetMember(ctx context.Context, groupChatID string, accountID string) (*query.Member, error) {
	panic(fmt.Errorf("not implemented: GetMember - getMember"))
}

// GetMembers is the resolver for the getMembers field.
func (r *queryRootResolver) GetMembers(ctx context.Context, groupChatID string, accountID string) ([]*query.Member, error) {
	panic(fmt.Errorf("not implemented: GetMembers - getMembers"))
}

// GetMessage is the resolver for the getMessage field.
func (r *queryRootResolver) GetMessage(ctx context.Context, messageID string, accountID string) (*query.Message, error) {
	panic(fmt.Errorf("not implemented: GetMessage - getMessage"))
}

// GetMessages is the resolver for the getMessages field.
func (r *queryRootResolver) GetMessages(ctx context.Context, groupChatID string, accountID string) ([]*query.Message, error) {
	panic(fmt.Errorf("not implemented: GetMessages - getMessages"))
}

// GroupChats is the resolver for the groupChats field.
func (r *subscriptionRootResolver) GroupChats(ctx context.Context, groupChatID string) (<-chan string, error) {
	panic(fmt.Errorf("not implemented: GroupChats - groupChats"))
}

// QueryRoot returns QueryRootResolver implementation.
func (r *Resolver) QueryRoot() QueryRootResolver { return &queryRootResolver{r} }

// SubscriptionRoot returns SubscriptionRootResolver implementation.
func (r *Resolver) SubscriptionRoot() SubscriptionRootResolver { return &subscriptionRootResolver{r} }

type queryRootResolver struct{ *Resolver }
type subscriptionRootResolver struct{ *Resolver }
