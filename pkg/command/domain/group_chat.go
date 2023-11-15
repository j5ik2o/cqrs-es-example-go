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

type GroupChat struct {
	id       *models.GroupChatId
	name     *models.GroupChatName
	members  *models.Members
	messages *models.Messages
	seqNr    uint64
	version  uint64
	deleted  bool
}

func ReplayGroupChat(events []esa.Event, snapshot *GroupChat) *GroupChat {
	result := snapshot
	for _, event := range events {
		result = result.ApplyEvent(event)
	}
	return result
}

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

func NewGroupChat(name *models.GroupChatName, administratorId *models.UserAccountId, executorId *models.UserAccountId) (*GroupChat, events.GroupChatEvent) {
	id := models.NewGroupChatId()
	members := models.NewMembers(administratorId)
	seqNr := uint64(1)
	version := uint64(1)
	return &GroupChat{id, name, members, models.NewMessages(), seqNr, version, false},
		events.NewGroupChatCreated(id, name, members, seqNr, executorId)
}

func NewGroupChatFrom(id *models.GroupChatId, name *models.GroupChatName, members *models.Members, messages *models.Messages, seqNr uint64, version uint64, deleted bool) *GroupChat {
	return &GroupChat{id, name, members, messages, seqNr, version, deleted}
}

func (g *GroupChat) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Id":       g.id.ToJSON(),
		"Name":     g.name.ToJSON(),
		"Members":  g.members.ToJSON(),
		"Messages": g.messages.ToJSON(),
		"SeqNr":    g.seqNr,
		"Version":  g.version,
		"Deleted":  g.deleted,
	}
}

func (g *GroupChat) GetId() esa.AggregateId {
	return g.id
}

func (g *GroupChat) GetGroupChatId() *models.GroupChatId {
	return g.id
}

func (g *GroupChat) GetName() *models.GroupChatName {
	return g.name
}

func (g *GroupChat) GetMembers() *models.Members {
	return g.members
}

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

func (g *GroupChat) IsDeleted() bool {
	return g.deleted
}

func (g *GroupChat) WithName(name *models.GroupChatName) *GroupChat {
	return NewGroupChatFrom(g.id, name, g.members, g.messages, g.seqNr, g.version, g.deleted)
}

func (g *GroupChat) WithMembers(members *models.Members) *GroupChat {
	return NewGroupChatFrom(g.id, g.name, members, g.messages, g.seqNr, g.version, g.deleted)
}

func (g *GroupChat) WithMessages(messages *models.Messages) *GroupChat {
	return NewGroupChatFrom(g.id, g.name, g.members, messages, g.seqNr, g.version, g.deleted)
}

func (g *GroupChat) WithVersion(version uint64) esa.Aggregate {
	return &GroupChat{id: g.id, seqNr: g.seqNr, version: version}
}

func (g *GroupChat) WithDeleted() *GroupChat {
	return NewGroupChatFrom(g.id, g.name, g.members, g.messages, g.seqNr, g.version, true)
}

func (g *GroupChat) AddMember(memberId *models.MemberId, userAccountId *models.UserAccountId, role models.Role, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("executorId is not the member of the group chat"))
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("userAccountId is already the member of the group chat"))
	}
	newMember := models.NewMember(memberId, userAccountId, role)
	newState := g.WithMembers(g.members.AddMember(userAccountId))
	newState.seqNr += 1
	memberAdded := events.NewGroupChatMemberAdded(newState.id, newMember, newState.seqNr, userAccountId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, memberAdded)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) RemoveMemberByUserAccountId(userAccountId *models.UserAccountId, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatRemoveMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatRemoveMemberErr("executorId is not the member of the group chat"))
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatRemoveMemberErr("userAccountId is already the member of the group chat"))
	}
	newState := g.WithMembers(g.members.RemoveMemberByUserAccountId(userAccountId))
	newState.seqNr += 1
	memberRemoved := events.NewGroupChatMemberRemoved(newState.id, userAccountId, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, memberRemoved)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) Rename(name *models.GroupChatName, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("executorId is not a newMember of the group chat"))
	}
	if g.name == name {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("name is already the same as the current name"))
	}
	newState := g.WithName(name)
	newState.seqNr += 1
	renamed := events.NewGroupChatRenamed(newState.id, name, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, renamed)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) Delete(executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteErr("executorId is not the member of the group chat"))
	}
	newState := g.WithDeleted()
	newState.seqNr += 1
	deleted := events.NewGroupChatDeleted(newState.id, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, deleted)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) PostMessage(message *models.Message, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("The group chat is deleted"))
	}
	if !g.members.IsMember(message.GetSenderId()) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("senderId is not the member of the group chat"))
	}
	if !g.members.IsMember(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("executorId is not the member of the group chat"))
	}
	if message.GetSenderId() != executorId {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("executorId is not the senderId of the message"))
	}
	newMessages, exists := g.messages.Add(message).Get()
	if !exists {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("message is already posted"))
	}
	newState := g.WithMessages(newMessages)
	newState.seqNr += 1
	messagePosted := events.NewGroupChatMessagePosted(newState.id, message, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, messagePosted)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) DeleteMessage(messageId *models.MessageId, executorId *models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteMessageErr("The group chat is deleted"))
	}
	if !g.members.IsMember(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatPostMessageErr("executorId is not the member of the group chat"))
	}
	message, exists := g.messages.Get(messageId).Get()
	if !exists {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteMessageErr("message is not found"))
	}
	member := g.members.FindByUserAccountId(message.GetSenderId()).MustGet()
	if member.GetUserAccountId() != executorId {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatDeleteMessageErr("User is not the sender of the message"))
	}
	newState := g.WithMessages(g.messages.Remove(messageId).MustGet())
	newState.seqNr += 1
	messageDeleted := events.NewGroupChatMessageDeleted(newState.id, messageId, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, messageDeleted)
	return mo.Ok(GroupChatWithEventPair(pair))
}
