package models

import (
	"github.com/samber/mo"
)

type Members struct {
	members []*Member
}

func NewMembers(administratorId *UserAccountId) *Members {
	return &Members{
		members: []*Member{
			NewMember(NewMemberId(), administratorId, AdminRole),
		},
	}
}

func (m *Members) AddMember(userAccountId *UserAccountId) *Members {
	newMembers := make([]*Member, len(m.members))
	copy(newMembers, m.members)
	newMembers = append(newMembers, NewMember(NewMemberId(), userAccountId, MemberRole))
	return &Members{
		members: newMembers,
	}
}

func (m *Members) RemoveMember(userAccountId *UserAccountId) *Members {
	newMembers := make([]*Member, 0, len(m.members))
	for _, member := range m.members {
		if member.userAccountId != userAccountId {
			newMembers = append(newMembers, member)
		}
	}
	return &Members{
		members: newMembers,
	}
}

func (m *Members) GetAdministrator() *Member {
	for _, member := range m.members {
		if member.role == AdminRole {
			return member
		}
	}
	return nil
}

func (m *Members) IsAdministrator(userAccountId *UserAccountId) bool {
	for _, member := range m.members {
		if member.userAccountId == userAccountId && member.role == AdminRole {
			return true
		}
	}
	return false
}

func (m *Members) IsMember(userAccountId *UserAccountId) bool {
	for _, member := range m.members {
		if member.userAccountId == userAccountId {
			return true
		}
	}
	return false
}

func (m *Members) FindByMemberId(memberId *MemberId) mo.Option[*Member] {
	for _, member := range m.members {
		if member.id == memberId {
			return mo.Some(member)
		}
	}
	return mo.None[*Member]()
}

func (m *Members) FindByUserAccountId(userAccountId *UserAccountId) mo.Option[*Member] {
	for _, member := range m.members {
		if member.userAccountId == userAccountId {
			return mo.Some(member)
		}
	}
	return mo.None[*Member]()
}

func (m *Members) ToArray() []*Member {
	copiedMembers := make([]*Member, len(m.members))
	copy(copiedMembers, m.members)
	return copiedMembers
}
