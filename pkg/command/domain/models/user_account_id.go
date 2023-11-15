package models

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

type UserAccountId struct {
	value string
}

func NewUserAccountId() *UserAccountId {
	id := ulid.Make()
	return &UserAccountId{value: id.String()}
}

func NewUserAccountIdFromString(value string) mo.Result[*UserAccountId] {
	if value == "" {
		return mo.Err[*UserAccountId](errors.New("UserAccountId is empty"))
	}
	return mo.Ok(&UserAccountId{value: value})
}

func ConvertUserAccountIdFromJSON(value map[string]interface{}) mo.Result[*UserAccountId] {
	return NewUserAccountIdFromString(value["Value"].(string))
}

func (u *UserAccountId) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"Value": u.value,
	}
}

func (u *UserAccountId) GetValue() string {
	return u.value
}

func (u *UserAccountId) GetTypeName() string {
	return "user-account"
}

func (u *UserAccountId) AsString() string {
	return fmt.Sprintf("%s-%s", u.GetTypeName(), u.GetValue())
}

func (u *UserAccountId) String() string {
	return u.AsString()
}
