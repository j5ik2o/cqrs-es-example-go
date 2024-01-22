package domain

import (
	"cqrs-es-example-go/pkg/command/domain/errors"
	"cqrs-es-example-go/pkg/command/domain/events"
	"cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	gt "github.com/barweiss/go-tuple"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/samber/mo"
)

// GroupChat is an aggregate of a group chat.
// GroupChat はグループチャットの集約です。
type GroupChat struct {
	id       *models.GroupChatId
	name     *models.GroupChatName
	members  *models.Members
	messages *models.Messages
	seqNr    uint64
	version  uint64
	deleted  bool
}

// ReplayGroupChat replays the events to the aggregate.
// ReplayGroupChat はイベントを集約にリプレイします。
func ReplayGroupChat(events []esa.Event, snapshot *GroupChat) *GroupChat {
	result := snapshot
	for _, event := range events {
		result = result.ApplyEvent(event)
	}
	return result
}

// ApplyEvent applies the event to the aggregate.
// ApplyEvent はイベントを集約に適用します。
func (g *GroupChat) ApplyEvent(event esa.Event) *GroupChat {
	switch e := event.(type) {
	case *events.GroupChatDeleted:
		result := g.Delete(e.GetExecutorId()).MustGet()
		return result.V1
	case *events.GroupChatMemberAdded:
		result := g.AddMember(e.GetMember().GetId(), e.GetMember().GetUserAccountId(), e.GetMember().GetRole(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events.GroupChatMemberRemoved:
		result := g.RemoveMemberByUserAccountId(e.GetUserAccountId(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events.GroupChatRenamed:
		result := g.Rename(e.GetName(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events.GroupChatMessagePosted:
		result := g.PostMessage(e.GetMessage(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events.GroupChatMessageDeleted:
		result := g.DeleteMessage(e.GetMessageId(), e.GetExecutorId()).MustGet()
		return result.V1
	default:
		return g
	}
}

// NewGroupChat creates a new group chat.
// NewGroupChat は新しいグループチャットを作成します。
func NewGroupChat(name *models.GroupChatName, executorId *models.UserAccountId) (*GroupChat, events.GroupChatEvent) {
	id := models.NewGroupChatId()
	members := models.NewMembers(executorId)
	seqNr := uint64(1)
	version := uint64(1)
	return &GroupChat{id, name, members, models.NewMessages(), seqNr, version, false},
		events.NewGroupChatCreated(id, name, members, seqNr, executorId)
}

// NewGroupChatFrom creates a new group chat from the specified parameters.
// NewGroupChatFrom は指定されたパラメータから新しいグループチャットを作成します。
func NewGroupChatFrom(id *models.GroupChatId, name *models.GroupChatName, members *models.Members, messages *models.Messages, seqNr uint64, version uint64, deleted bool) *GroupChat {
	return &GroupChat{id, name, members, messages, seqNr, version, deleted}
}

// ToJSON converts the aggregate to JSON.
// ToJSON は集約を JSON に変換します。
//
// However, this method is out of layer.
// ただし、このメソッドはレイヤーを逸脱しています。
func (g *GroupChat) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":       g.id.ToJSON(),
		"name":     g.name.ToJSON(),
		"members":  g.members.ToJSON(),
		"messages": g.messages.ToJSON(),
		"seq_nr":   g.seqNr,
		"version":  g.version,
		"deleted":  g.deleted,
	}
}

// GetId returns the aggregate GetId.
// GetId は集約の GetId を返します。
func (g *GroupChat) GetId() esa.AggregateId {
	return g.id
}

// GetGroupChatId returns the aggregate GetGroupChatId.
// GetGroupChatId は集約の GetGroupChatId を返します。
func (g *GroupChat) GetGroupChatId() *models.GroupChatId {
	return g.id
}

// GetName returns the aggregate GetName.
// GetName は集約の GetName を返します。
func (g *GroupChat) GetName() *models.GroupChatName {
	return g.name
}

// GetMembers returns the aggregate GetMembers.
// GetMembers は集約の GetMembers を返します。
func (g *GroupChat) GetMembers() *models.Members {
	return g.members
}

// GetMessages returns the aggregate GetMessages.
// GetMessages は集約の GetMessages を返します。
func (g *GroupChat) GetMessages() *models.Messages {
	return g.messages
}

func (g *GroupChat) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChat) GetVersion() uint64 {
	return g.version
}

func (g *GroupChat) String() string {
	return fmt.Sprintf("id: %s, seqNr: %d, version: %d", g.id, g.seqNr, g.version)
}

// IsDeleted returns whether the aggregate is deleted.
// IsDeleted は集約が削除されたかどうかを返します。
//
// # Returns / 戻り値:
// - true if the aggregate is deleted / 集約が削除された場合は true
func (g *GroupChat) IsDeleted() bool {
	return g.deleted
}

// WithName returns a new aggregate with the specified name.
// WithName は指定された名前の新しい集約を返します。
//
// # Returns / 戻り値:
// - The new aggregate / 新しい集約
func (g *GroupChat) WithName(name *models.GroupChatName) *GroupChat {
	return NewGroupChatFrom(g.id, name, g.members, g.messages, g.seqNr, g.version, g.deleted)
}

// WithMembers returns a new aggregate with the specified members.
// WithMembers は指定されたメンバーの新しい集約を返します。
//
// # Returns / 戻り値:
// - The new aggregate / 新しい集約
func (g *GroupChat) WithMembers(members *models.Members) *GroupChat {
	return NewGroupChatFrom(g.id, g.name, members, g.messages, g.seqNr, g.version, g.deleted)
}

// WithMessages returns a new aggregate with the specified messages.
// WithMessages は指定されたメッセージの新しい集約を返します。
//
// # Returns / 戻り値:
// - The new aggregate / 新しい集約
func (g *GroupChat) WithMessages(messages *models.Messages) *GroupChat {
	return NewGroupChatFrom(g.id, g.name, g.members, messages, g.seqNr, g.version, g.deleted)
}

// WithVersion returns a new aggregate with the specified version.
// WithVersion は指定されたバージョンの新しい集約を返します。
//
// # Returns / 戻り値:
// - The new aggregate / 新しい集約
func (g *GroupChat) WithVersion(version uint64) esa.Aggregate {
	return NewGroupChatFrom(g.id, g.name, g.members, g.messages, g.seqNr, version, g.deleted)
}

// WithDeleted returns a new aggregate with the deleted flag.
// WithDeleted は削除フラグのある新しい集約を返します。
//
// # Returns / 戻り値:
// - The new aggregate / 新しい集約
func (g *GroupChat) WithDeleted() *GroupChat {
	return NewGroupChatFrom(g.id, g.name, g.members, g.messages, g.seqNr, g.version, true)
}

// AddMember adds a new member to the aggregate.
// AddMember は新しいメンバーを集約に追加します。
//
// # Parameters / 引数:
// - memberId: The member ID to be assigned / 割り当てるメンバーID
// - userAccountId: The user account ID of the member / メンバーのユーザーアカウントID
// - role: The role of the member / メンバーの役割
// - executorId: The user account ID of the executor / 実行者のユーザーアカウントID
// # Constraints / 制約:
// - The group chat is not deleted / グループチャットが削除されていないこと
// - The userAccountId is not the member of the group chat / userAccountId がグループチャットのメンバーでないこと
// - The executorId is the administrator of the group chat / executorId がグループチャットの管理者であること
// # Returns / 戻り値:
// - The result of the operation / 操作の結果
func (g *GroupChat) AddMember(
	memberId *models.MemberId,
	userAccountId *models.UserAccountId,
	role models.Role,
	executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The userAccountId is already the member of the group chat"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The executorId is not the member of the group chat"))
	}
	newMember := models.NewMember(memberId, userAccountId, role)
	newState := g.WithMembers(g.members.AddMember(userAccountId))
	newState.seqNr += 1
	memberAdded := events.NewGroupChatMemberAdded(newState.id, newMember, newState.seqNr, userAccountId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, memberAdded)
	return mo.Ok(GroupChatWithEventPair(pair))
}

// RemoveMemberByUserAccountId removes the member from the aggregate.
// RemoveMemberByUserAccountId はメンバーを集約から削除します。
//
// # Parameters / 引数:
// - userAccountId: The user account ID of the member / メンバーのユーザーアカウントID
// - executorId: The user account ID of the executor / 実行者のユーザーアカウントID
// # Constraints / 制約:
// - The group chat is not deleted / グループチャットが削除されていないこと
// - The userAccountId is the administrator of the group chat / userAccountId がグループチャットのメンバーであること
// - The executorId is the administrator of the group chat / executorId がグループチャットの管理者であること
// # Returns / 戻り値:
// - The result of the operation / 操作の結果
func (g *GroupChat) RemoveMemberByUserAccountId(userAccountId *models.UserAccountId, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatRemoveMemberErr("The group chat is deleted"))
	}
	if !g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatRemoveMemberErr("The userAccountId is not the member of the group chat"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatRemoveMemberErr("The executorId is not the administrator of the group chat"))
	}
	newState := g.WithMembers(g.members.RemoveMemberByUserAccountId(userAccountId))
	newState.seqNr += 1
	memberRemoved := events.NewGroupChatMemberRemoved(newState.id, userAccountId, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, memberRemoved)
	return mo.Ok(GroupChatWithEventPair(pair))
}

// Rename renames the aggregate.
// Rename は集約の名前を変更します。
//
// # Parameters / 引数:
// - name: The new name of the aggregate / 集約の新しい名前
// - executorId: The user account ID of the executor / 実行者のユーザーアカウントID
// # Constraints / 制約:
// - The group chat is not deleted / グループチャットが削除されていないこと
// - The executorId is the administrator of the group chat / executorId がグループチャットの管理者であること
// - The name is not the same as the current name / name が現在の名前と同じでないこと
// # Returns / 戻り値:
// - The result of the operation / 操作の結果
func (g *GroupChat) Rename(name *models.GroupChatName, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The executorId is not an administrator of the group chat"))
	}
	if g.name == name {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The name is already the same as the current name"))
	}
	newState := g.WithName(name)
	newState.seqNr += 1
	renamed := events.NewGroupChatRenamed(newState.id, name, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, renamed)
	return mo.Ok(GroupChatWithEventPair(pair))
}

// Delete deletes the aggregate.
// Delete は集約を削除します。
//
// # Parameters / 引数:
// - executorId: The user account ID of the executor / 実行者のユーザーアカウントID
// # Constraints / 制約:
// - The group chat is not deleted / グループチャットが削除されていないこと
// - The executorId is the administrator of the group chat / executorId がグループチャットの管理者であること
// # Returns / 戻り値:
// - The result of the operation / 操作の結果
func (g *GroupChat) Delete(executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteErr("The executorId is not the member of the group chat"))
	}
	newState := g.WithDeleted()
	newState.seqNr += 1
	deleted := events.NewGroupChatDeleted(newState.id, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, deleted)
	return mo.Ok(GroupChatWithEventPair(pair))
}

// PostMessage posts a new message to the aggregate.
// PostMessage は新しいメッセージを集約に投稿します。
//
// # Parameters / 引数:
// - message: The message to be posted / 投稿するメッセージ
// - executorId: The user account ID of the executor / 実行者のユーザーアカウントID
// # Constraints / 制約:
// - The group chat is not deleted / グループチャットが削除されていないこと
// - The Message#senderId is the member of the group chat / senderId がグループチャットのメンバーであること
// - The executorId is the senderId of the message / executorId がメッセージの senderId であること
// - The message is not already posted / メッセージがすでに投稿されていないこと
// # Returns / 戻り値:
// - The result of the operation / 操作の結果
func (g *GroupChat) PostMessage(message *models.Message, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The group chat is deleted"))
	}
	if !g.members.IsMember(message.GetSenderId()) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The senderId is not the member of the group chat"))
	}
	if !g.members.IsMember(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The executorId is not the member of the group chat"))
	}
	if !message.GetSenderId().Equals(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The executorId is not the senderId of the message"))
	}
	newMessages, exists := g.messages.Add(message).Get()
	if !exists {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The message is already posted"))
	}
	newState := g.WithMessages(newMessages)
	newState.seqNr += 1
	messagePosted := events.NewGroupChatMessagePosted(newState.id, message, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, messagePosted)
	return mo.Ok(GroupChatWithEventPair(pair))
}

// DeleteMessage deletes the message from the aggregate.
// DeleteMessage はメッセージを集約から削除します。
//
// # Parameters / 引数:
// - messageId: The ID of the message to be deleted / 削除するメッセージのID
// - executorId: The user account ID of the executor / 実行者のユーザーアカウントID
// # Constraints / 制約:
// - The group chat is not deleted / グループチャットが削除されていないこと
// - The executorId is the sender of the message / executorId がメッセージの senderId であること
// - The message is not already deleted / メッセージがすでに削除されていないこと
// # Returns / 戻り値:
// - The result of the operation / 操作の結果
func (g *GroupChat) DeleteMessage(messageId *models.MessageId, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteMessageErr("The group chat is deleted"))
	}
	if !g.members.IsMember(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The executorId is not the member of the group chat"))
	}
	message, exists := g.messages.Get(messageId).Get()
	if !exists {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteMessageErr("The message is not found"))
	}
	member := g.members.FindByUserAccountId(message.GetSenderId()).MustGet()
	if !member.GetUserAccountId().Equals(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteMessageErr("The executorId is not the sender of the message"))
	}
	newState := g.WithMessages(g.messages.Remove(messageId).MustGet())
	newState.seqNr += 1
	messageDeleted := events.NewGroupChatMessageDeleted(newState.id, messageId, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, messageDeleted)
	return mo.Ok(GroupChatWithEventPair(pair))
}
