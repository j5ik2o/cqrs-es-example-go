package models

import (
	"github.com/samber/mo"
)

// Members is the fist-class collection of Member.
type Members struct {
	values []*Member
}

// NewMembers is the constructor for Members.
func NewMembers(administratorId UserAccountId) *Members {
	return &Members{
		values: []*Member{
			NewMember(NewMemberId(), administratorId, AdminRole),
		},
	}
}

// NewMembersFrom is the constructor for Members.
func NewMembersFrom(values []*Member) *Members {
	return &Members{
		values: values,
	}
}

// ConvertMembersFromJSON is a constructor for Members.
func ConvertMembersFromJSON(value map[string]interface{}) *Members {
	values := value["values"].([]interface{})
	members := make([]*Member, len(values))
	for i, v := range values {
		members[i] = ConvertMemberFromJSON(v.(map[string]interface{}))
	}
	return NewMembersFrom(members)
}

// ToJSON converts to JSON.
func (m *Members) ToJSON() map[string]interface{} {
	values := make([]map[string]interface{}, len(m.values))
	for i, v := range m.values {
		values[i] = v.ToJSON()
	}
	return map[string]interface{}{
		"values": values,
	}
}

// AddMember adds a member.
func (m *Members) AddMember(userAccountId UserAccountId) *Members {
	newMembers := make([]*Member, len(m.values))
	copy(newMembers, m.values)
	newMembers = append(newMembers, NewMember(NewMemberId(), userAccountId, MemberRole))
	return &Members{
		values: newMembers,
	}
}

// RemoveMemberByUserAccountId removes a member.
func (m *Members) RemoveMemberByUserAccountId(userAccountId *UserAccountId) *Members {
	newMembers := make([]*Member, 0, len(m.values))
	for _, member := range m.values {
		if !member.userAccountId.Equals(userAccountId) {
			newMembers = append(newMembers, member)
		}
	}
	return &Members{
		values: newMembers,
	}
}

// GetAdministrator returns administrator.
func (m *Members) GetAdministrator() *Member {
	for _, member := range m.values {
		if member.role == AdminRole {
			return member
		}
	}
	return nil
}

// IsAdministrator checks if the user is administrator.
func (m *Members) IsAdministrator(userAccountId *UserAccountId) bool {
	for _, member := range m.values {
		if member.userAccountId.Equals(userAccountId) && member.role == AdminRole {
			return true
		}
	}
	return false
}

// IsMember checks if the user is member.
func (m *Members) IsMember(userAccountId *UserAccountId) bool {
	for _, member := range m.values {
		if member.userAccountId.Equals(userAccountId) {
			return true
		}
	}
	return false
}

// FindByMemberId finds a member by member id.
func (m *Members) FindByMemberId(memberId *MemberId) mo.Option[*Member] {
	for _, member := range m.values {
		if member.id.Equals(memberId) {
			return mo.Some(member)
		}
	}
	return mo.None[*Member]()
}

// FindByUserAccountId finds a member by user account id.
func (m *Members) FindByUserAccountId(userAccountId *UserAccountId) mo.Option[*Member] {
	for _, member := range m.values {
		if member.userAccountId.Equals(userAccountId) {
			return mo.Some(member)
		}
	}
	return mo.None[*Member]()
}

// ToSlice converts to slice.
func (m *Members) ToSlice() []*Member {
	copiedMembers := make([]*Member, len(m.values))
	copy(copiedMembers, m.values)
	return copiedMembers
}
