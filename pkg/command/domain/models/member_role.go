package models

import (
	"fmt"
	"strings"
)

type Role int

const (
	MemberRole = iota
	AdminRole
)

func (r Role) String() string {
	switch r {
	case MemberRole:
		return "member"
	case AdminRole:
		return "admin"
	default:
		panic("unknown role")
	}
}

func StringToRole(s string) (Role, error) {
	switch strings.ToLower(s) {
	case "member":
		return MemberRole, nil
	case "admin":
		return AdminRole, nil
	default:
		return 0, fmt.Errorf("unknown role: %s", s)
	}
}
