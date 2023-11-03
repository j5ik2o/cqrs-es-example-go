package domain

import (
	"cqrs-es-example-go/domain/errors"
	"cqrs-es-example-go/domain/events"
	"cqrs-es-example-go/domain/models"
	"fmt"
	gt "github.com/barweiss/go-tuple"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/samber/mo"
)

type GroupChat struct {
	id      models.GroupChatId
	seqNr   uint64
	version uint64
	name    models.GroupChatName
	members models.Members
	deleted bool
}

func NewGroupChat(name models.GroupChatName, members models.Members) *GroupChat {
	id := models.NewGroupChatId()
	seqNr := uint64(1)
	version := uint64(1)
	return &GroupChat{id, seqNr, version, name, members, false}
}

func NewGroupChatFrom(id models.GroupChatId, seqNr uint64, version uint64, name models.GroupChatName, members models.Members, deleted bool) *GroupChat {
	return &GroupChat{id, seqNr, version, name, members, deleted}
}

func (g *GroupChat) GetId() esa.AggregateId {
	return &g.id
}

func (g *GroupChat) GetSeqNr() uint64 {
	return g.seqNr
}

func (g *GroupChat) GetVersion() uint64 {
	return g.version
}

func (g *GroupChat) WithVersion(version uint64) esa.Aggregate {
	return &GroupChat{id: g.id, seqNr: g.seqNr, version: version}
}

func (g *GroupChat) String() string {
	return fmt.Sprintf("id: %s, seqNr: %d, version: %d", g.id, g.seqNr, g.version)
}

func (g *GroupChat) AddMember(memberId models.MemberId, userAccountId models.UserAccountId, role models.Role, executorId models.UserAccountId) mo.Result[GroupChatWithEventPair] {
	if g.deleted {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("The group chat is deleted"))
	}
	if !g.members.IsAdministrator(executorId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("executorId is not a newMember of the group chat"))
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[GroupChatWithEventPair](errors.NewGroupChatAddMemberErr("userAccountId is already a newMember of the group chat"))
	}
	newMember := models.NewMember(memberId, userAccountId, role)
	newState := NewGroupChatFrom(g.id, g.seqNr+1, g.version+1, g.name, *g.members.AddMember(userAccountId), g.deleted)
	memberAdded := events.NewGroupChatMemberAdded(g.id, newState.seqNr, *newMember, userAccountId)
	pair := gt.New2[*GroupChat, events.GroupChatEvent](newState, memberAdded)
	return mo.Ok[GroupChatWithEventPair](GroupChatWithEventPair(pair))
}

func (g *GroupChat) GetMembers() *models.Members {
	return &g.members
}
