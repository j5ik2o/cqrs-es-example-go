package models

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"
)

const UserAccountIdPrefix = "UserAccount"

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
	if len(value) > len(UserAccountIdPrefix) && value[0:len(UserAccountIdPrefix)] == UserAccountIdPrefix {
		value = value[len(UserAccountIdPrefix)+1:]
	}
	// 先頭がUserAccount-であれば、それを削除する
	if len(value) > 12 && value[0:12] == "UserAccount" {
		value = value[13:]
	}
	return mo.Ok(&UserAccountId{value: value})
}

func ConvertUserAccountIdFromJSON(value map[string]interface{}) mo.Result[*UserAccountId] {
	return NewUserAccountIdFromString(value["value"].(string))
}

func (u *UserAccountId) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"value": u.value,
	}
}

func (u *UserAccountId) GetValue() string {
	return u.value
}

func (u *UserAccountId) GetTypeName() string {
	return "UserAccount"
}

func (u *UserAccountId) AsString() string {
	return fmt.Sprintf("%s-%s", u.GetTypeName(), u.GetValue())
}

func (u *UserAccountId) String() string {
	return u.AsString()
}

func (u *UserAccountId) Equals(other *UserAccountId) bool {
	return u.value == other.value
}
