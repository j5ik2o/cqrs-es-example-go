package domain

type Member struct {
	id            MemberId
	userAccountId UserAccountId
	role          Role
}

func NewMember(id MemberId, userAccountId UserAccountId, role Role) *Member {
	return &Member{
		id:            id,
		userAccountId: userAccountId,
		role:          role,
	}
}
