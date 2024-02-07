package models

import (
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

// MemberId is a value object that represents a member id.
type MemberId struct {
	value string
}

// NewMemberId is the constructor for MemberId with generating id.
func NewMemberId() MemberId {
	id := ulid.Make()
	return MemberId{value: id.String()}
}

// NewMemberIdFromString is the constructor for MemberId.
func NewMemberIdFromString(value string) mo.Result[MemberId] {
	return mo.Ok(MemberId{value: value})
}

// ConvertMemberIdFromJSON is a constructor for MemberId.
func ConvertMemberIdFromJSON(value map[string]interface{}) mo.Result[MemberId] {
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

func (m *MemberId) Equals(other *MemberId) bool {
	return m.value == other.value
}
