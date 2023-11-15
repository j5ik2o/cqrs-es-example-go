package models

import (
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

type MemberId struct {
	value string
}

func NewMemberId() *MemberId {
	id := ulid.Make()
	return &MemberId{value: id.String()}
}

func NewMemberIdFromString(value string) mo.Result[*MemberId] {
	return mo.Ok(&MemberId{value: value})
}

func ConvertMemberIdFromJSON(value map[string]interface{}) mo.Result[*MemberId] {
	return NewMemberIdFromString(value["value"].(string))
}

func (m *MemberId) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func (m *MemberId) GetValue() string {
	return m.value
}

func (m *MemberId) String() string {
	return m.value
}
