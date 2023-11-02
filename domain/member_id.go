package domain

import "github.com/oklog/ulid/v2"

type MemberId struct {
	value string
}

func NewMemberId() *MemberId {
	id := ulid.Make()
	return &MemberId{value: id.String()}
}

func NewMemberIdFromString(value string) *MemberId {
	return &MemberId{value: value}
}

func (m *MemberId) GetValue() string {
	return m.value
}

func (m *MemberId) String() string {
	return m.value
}
