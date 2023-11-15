package domain

import (
	errors2 "cqrs-es-example-go/pkg/command/domain/errors"
	events2 "cqrs-es-example-go/pkg/command/domain/events"
	models2 "cqrs-es-example-go/pkg/command/domain/models"
	"fmt"
	gt "github.com/barweiss/go-tuple"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/samber/mo"
)

type GroupChat struct {
	id       *models2.GroupChatId
	name     *models2.GroupChatName
	members  *models2.Members
	messages *models2.Messages
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
	case *events2.GroupChatDeleted:
		result := g.Delete(e.GetExecutorId()).MustGet()
		return result.V1
	case *events2.GroupChatMemberAdded:
		result := g.AddMember(e.GetMember().GetId(), e.GetMember().GetUserAccountId(), e.GetMember().GetRole(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events2.GroupChatMemberRemoved:
		result := g.RemoveMemberByUserAccountId(e.GetUserAccountId(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events2.GroupChatRenamed:
		result := g.Rename(e.GetName(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events2.GroupChatMessagePosted:
		result := g.PostMessage(e.GetMessage(), e.GetExecutorId()).MustGet()
		return result.V1
	case *events2.GroupChatMessageDeleted:
		result := g.DeleteMessage(e.GetMessageId(), e.GetExecutorId()).MustGet()
		return result.V1
	default:
		return g
	}
}

func NewGroupChat(name *models2.GroupChatName, administratorId *models2.UserAccountId, executorId *models2.UserAccountId) (*GroupChat, events2.GroupChatEvent) {
	id := models2.NewGroupChatId()
	members := models2.NewMembers(administratorId)
	seqNr := uint64(1)
	version := uint64(1)
	return &GroupChat{id, name, members, models2.NewMessages(), seqNr, version, false},
		events2.NewGroupChatCreated(id, name, members, seqNr, executorId)
}

func NewGroupChatFrom(id *models2.GroupChatId, name *models2.GroupChatName, members *models2.Members, messages *models2.Messages, seqNr uint64, version uint64, deleted bool) *GroupChat {
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

func (g *GroupChat) GetGroupChatId() *models2.GroupChatId {
	return g.id
}

func (g *GroupChat) GetName() *models2.GroupChatName {
	return g.name
}

func (g *GroupChat) GetMembers() *models2.Members {
	return g.members
}

func (g *GroupChat) GetMessages() *models2.Messages {
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

func (g *GroupChat) WithName(name *models2.GroupChatName) *GroupChat {
	return NewGroupChatFrom(g.id, name, g.members, g.messages, g.seqNr, g.version, g.deleted)
}

func (g *GroupChat) WithMembers(members *models2.Members) *GroupChat {
	return NewGroupChatFrom(g.id, g.name, members, g.messages, g.seqNr, g.version, g.deleted)
}

func (g *GroupChat) WithMessages(messages *models2.Messages) *GroupChat {
	return NewGroupChatFrom(g.id, g.name, g.members, messages, g.seqNr, g.version, g.deleted)
}

func (g *GroupChat) WithVersion(version uint64) esa.Aggregate {
	return &GroupChat{id: g.id, seqNr: g.seqNr, version: version}
}

func (g *GroupChat) WithDeleted() *GroupChat {
	return NewGroupChatFrom(g.id, g.name, g.members, g.messages, g.seqNr, g.version, true)
}

func (g *GroupChat) AddMember(memberId *models2.MemberId, userAccountId *models2.UserAccountId, role models2.Role, executorId *models2.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatAddMemberErr("executorId is not the member of the group chat"))
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatAddMemberErr("userAccountId is already the member of the group chat"))
	}
	newMember := models2.NewMember(memberId, userAccountId, role)
	newState := g.WithMembers(g.members.AddMember(userAccountId))
	newState.seqNr += 1
	memberAdded := events2.NewGroupChatMemberAdded(newState.id, newMember, newState.seqNr, userAccountId)
	pair := gt.New2[*GroupChat, events2.GroupChatEvent](newState, memberAdded)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) RemoveMemberByUserAccountId(userAccountId *models2.UserAccountId, executorId *models2.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatRemoveMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatRemoveMemberErr("executorId is not the member of the group chat"))
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatRemoveMemberErr("userAccountId is already the member of the group chat"))
	}
	newState := g.WithMembers(g.members.RemoveMemberByUserAccountId(userAccountId))
	newState.seqNr += 1
	memberRemoved := events2.NewGroupChatMemberRemoved(newState.id, userAccountId, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events2.GroupChatEvent](newState, memberRemoved)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) Rename(name *models2.GroupChatName, executorId *models2.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatAddMemberErr("executorId is not a newMember of the group chat"))
	}
	if g.name == name {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatAddMemberErr("name is already the same as the current name"))
	}
	newState := g.WithName(name)
	newState.seqNr += 1
	renamed := events2.NewGroupChatRenamed(newState.id, name, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events2.GroupChatEvent](newState, renamed)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) Delete(executorId *models2.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatDeleteErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatDeleteErr("executorId is not the member of the group chat"))
	}
	newState := g.WithDeleted()
	newState.seqNr += 1
	deleted := events2.NewGroupChatDeleted(newState.id, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events2.GroupChatEvent](newState, deleted)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) PostMessage(message *models2.Message, executorId *models2.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatPostMessageErr("The group chat is deleted"))
	}
	if !g.members.IsMember(message.GetSenderId()) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatPostMessageErr("senderId is not the member of the group chat"))
	}
	if !g.members.IsMember(executorId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatPostMessageErr("executorId is not the member of the group chat"))
	}
	if message.GetSenderId() != executorId {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatPostMessageErr("executorId is not the senderId of the message"))
	}
	newMessages, exists := g.messages.Add(message).Get()
	if !exists {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatPostMessageErr("message is already posted"))
	}
	newState := g.WithMessages(newMessages)
	newState.seqNr += 1
	messagePosted := events2.NewGroupChatMessagePosted(newState.id, message, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events2.GroupChatEvent](newState, messagePosted)
	return mo.Ok(GroupChatWithEventPair(pair))
}

func (g *GroupChat) DeleteMessage(messageId *models2.MessageId, executorId *models2.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatDeleteMessageErr("The group chat is deleted"))
	}
	if !g.members.IsMember(executorId) {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatPostMessageErr("executorId is not the member of the group chat"))
	}
	message, exists := g.messages.Get(messageId).Get()
	if !exists {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatDeleteMessageErr("message is not found"))
	}
	member := g.members.FindByUserAccountId(message.GetSenderId()).MustGet()
	if member.GetUserAccountId() != executorId {
		return mo.Err[GroupChatWithEventPair](errors2.NewGroupChatDeleteMessageErr("User is not the sender of the message"))
	}
	newState := g.WithMessages(g.messages.Remove(messageId).MustGet())
	newState.seqNr += 1
	messageDeleted := events2.NewGroupChatMessageDeleted(newState.id, messageId, newState.seqNr, executorId)
	pair := gt.New2[*GroupChat, events2.GroupChatEvent](newState, messageDeleted)
	return mo.Ok(GroupChatWithEventPair(pair))
}
