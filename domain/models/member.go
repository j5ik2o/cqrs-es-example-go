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

func (m *Member) GetId() *MemberId {
	return m.id
}

func (m *Member) GetUserAccountId() *UserAccountId {
	return m.userAccountId
}

func (m *Member) GetRole() Role {
	return m.role
}
