package domain

import (
	"cqrs-es-example-go/domain/events"
	"fmt"
	gt "github.com/barweiss/go-tuple"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/samber/mo"
)

type GroupChat struct {
	id      GroupChatId
	seqNr   uint64
	version uint64
	members Members
	deleted bool
}

func NewGroupChat(id GroupChatId, seqNr uint64, version uint64, members Members, deleted bool) *GroupChat {
	return &GroupChat{id, seqNr, version, members, deleted}
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

type GroupChatErr struct {
	Message string
}

func (e *GroupChatErr) Error() string {
	return e.Message
}

type GroupChatDeletedErr struct {
	GroupChatErr
}

func NewGroupChatDeletedErr(message string) *GroupChatDeletedErr {
	return &GroupChatDeletedErr{GroupChatErr{message}}
}

func (e *GroupChatDeletedErr) Error() string {
	return e.Message
}

func (g *GroupChat) AddMember(memberId MemberId, userAccountId UserAccountId, role Role) mo.Result[gt.Pair[*GroupChat, events.GroupChatEvent]] {
	if g.deleted {
		return mo.Err[gt.Pair[*GroupChat, events.GroupChatEvent]](&GroupChatDeletedErr{GroupChatErr{"The group chat is deleted"}})
	}
	if !g.members.IsAdministrator(userAccountId) {
		return mo.Err[gt.Pair[*GroupChat, events.GroupChatEvent]](&GroupChatErr{"executor_id is not a member of the group chat"})
	}
	if g.members.IsMember(userAccountId) {
		return mo.Err[gt.Pair[*GroupChat, events.GroupChatEvent]](&GroupChatErr{"executor_id is already a member of the group chat"})
	}
	member := NewMember(memberId, userAccountId, role)
	newState := NewGroupChat(g.id, g.seqNr+1, g.version+1, *g.members.AddMember(userAccountId), g.deleted)
	event := events.NewGroupChatMemberAdded(g.id, newState.seqNr, *member, userAccountId)
	t := gt.New2[*GroupChat, events.GroupChatEvent](newState, event)
	return mo.Ok[gt.Pair[*GroupChat, events.GroupChatEvent]](gt.Pair[*GroupChat, events.GroupChatEvent](t))
}
