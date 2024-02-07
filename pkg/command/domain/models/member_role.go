package models

import "strings"

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

func StringToRole(s string) Role {
	switch strings.ToLower(s) {
	case "member":
		return MemberRole
	case "admin":
		return AdminRole
	default:
		panic("unknown role string: " + s)
	}
}
