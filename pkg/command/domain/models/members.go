package models

import (
	"github.com/samber/mo"
)

type Members struct {
	values []*Member
}

func NewMembers(administratorId *UserAccountId) *Members {
	return &Members{
		values: []*Member{
			NewMember(NewMemberId(), administratorId, AdminRole),
		},
	}
}

func NewMembersFrom(values []*Member) *Members {
	return &Members{
		values: values,
	}
}

func ConvertMembersFromJSON(value map[string]interface{}) *Members {
	values := value["values"].([]interface{})
	members := make([]*Member, len(values))
	for i, v := range values {
		members[i] = ConvertMemberFromJSON(v.(map[string]interface{}))
	}
	return NewMembersFrom(members)
}

func (m *Members) ToJSON() map[string]interface{} {
	values := make([]map[string]interface{}, len(m.values))
	for i, v := range m.values {
		values[i] = v.ToJSON()
	}
	return map[string]interface{}{
		"values": values,
	}
}

func (m *Members) AddMember(userAccountId *UserAccountId) *Members {
	newMembers := make([]*Member, len(m.values))
	copy(newMembers, m.values)
	newMembers = append(newMembers, NewMember(NewMemberId(), userAccountId, MemberRole))
	return &Members{
		values: newMembers,
	}
}

func (m *Members) RemoveMemberByUserAccountId(userAccountId *UserAccountId) *Members {
	newMembers := make([]*Member, 0, len(m.values))
	for _, member := range m.values {
		if member.userAccountId != userAccountId {
			newMembers = append(newMembers, member)
		}
	}
	return &Members{
		values: newMembers,
	}
}

func (m *Members) GetAdministrator() *Member {
	for _, member := range m.values {
		if member.role == AdminRole {
			return member
		}
	}
	return nil
}

func (m *Members) IsAdministrator(userAccountId *UserAccountId) bool {
	for _, member := range m.values {
		if member.userAccountId == userAccountId && member.role == AdminRole {
			return true
		}
	}
	return false
}

func (m *Members) IsMember(userAccountId *UserAccountId) bool {
	for _, member := range m.values {
		if member.userAccountId == userAccountId {
			return true
		}
	}
	return false
}

func (m *Members) FindByMemberId(memberId *MemberId) mo.Option[*Member] {
	for _, member := range m.values {
		if member.id == memberId {
			return mo.Some(member)
		}
	}
	return mo.None[*Member]()
}

func (m *Members) FindByUserAccountId(userAccountId *UserAccountId) mo.Option[*Member] {
	for _, member := range m.values {
		if member.userAccountId == userAccountId {
			return mo.Some(member)
		}
	}
	return mo.None[*Member]()
}

func (m *Members) ToSlice() []*Member {
	copiedMembers := make([]*Member, len(m.values))
	copy(copiedMembers, m.values)
	return copiedMembers
}
