package models

type Member struct {
	id            *MemberId
	userAccountId *UserAccountId
	role          Role
}

func NewMember(id *MemberId, userAccountId *UserAccountId, role Role) *Member {
	return &Member{
		id:            id,
		userAccountId: userAccountId,
		role:          role,
	}
}

func ConvertMemberFromJSON(value map[string]interface{}) *Member {
	roleValue := value["Role"]
	role := Role(roleValue.(float64))
	return NewMember(
		ConvertMemberIdFromJSON(value["Id"].(map[string]interface{})).MustGet(),
		ConvertUserAccountIdFromJSON(value["UserAccountId"].(map[string]interface{})).MustGet(),
		role,
	)
}

func (m *Member) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Id":            m.id.ToJSON(),
		"UserAccountId": m.userAccountId.ToJSON(),
		"Role":          m.role,
	}
}

func (m *Member) GetId() *MemberId {
	return m.id
}

func (m *Member) GetUserAccountId() *UserAccountId {
	return m.userAccountId
}

func (m *Member) GetRole() Role {
	return m.role
}
