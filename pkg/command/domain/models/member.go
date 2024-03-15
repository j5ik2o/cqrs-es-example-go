package models

// Member is the model of member.
type Member struct {
	id            MemberId
	userAccountId UserAccountId
	role          Role
}

// NewMember is the constructor of Member.
func NewMember(id MemberId, userAccountId UserAccountId, role Role) Member {
	return Member{
		id:            id,
		userAccountId: userAccountId,
		role:          role,
	}
}

// ConvertMemberFromJSON converts JSON to Member.
func ConvertMemberFromJSON(value map[string]interface{}) Member {
	roleValue := value["role"]
	role := Role(roleValue.(float64))
	return NewMember(
		ConvertMemberIdFromJSON(value["id"].(map[string]interface{})).MustGet(),
		ConvertUserAccountIdFromJSON(value["user_account_id"].(map[string]interface{})).MustGet(),
		role,
	)
}

func (m *Member) Equals(other *Member) bool {
	return m.id.Equals(&other.id) && m.userAccountId.Equals(&other.userAccountId) && m.role == other.role
}

// ToJSON converts to JSON.
//
// However, this method is out of layer.
func (m *Member) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":              m.id.ToJSON(),
		"user_account_id": m.userAccountId.ToJSON(),
		"role":            m.role,
	}
}

// GetId returns id.
func (m *Member) GetId() *MemberId {
	return &m.id
}

// GetUserAccountId returns user account id.
func (m *Member) GetUserAccountId() *UserAccountId {
	return &m.userAccountId
}

// GetRole returns role.
func (m *Member) GetRole() Role {
	return m.role
}
