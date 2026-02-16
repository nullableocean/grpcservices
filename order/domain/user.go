package domain

import pkg "main/pkg/roles"

type User struct {
	id       int64
	username string
	roles    []pkg.UserRole
}

func NewUser(id int64) *User {
	return &User{
		id:    id,
		roles: make([]pkg.UserRole, 0),
	}
}

func (u *User) Id() int64 {
	return u.id
}

func (u *User) Username() string {
	return u.username
}

func (u *User) Roles() []pkg.UserRole {
	return u.roles
}

func (u *User) SetUsername(username string) {
	u.username = username
}

func (u *User) SetRoles(roles []pkg.UserRole) {
	u.roles = roles
}

func (u *User) AddRole(role pkg.UserRole) {
	u.roles = append(u.roles, role)
}

func (u *User) HasRole(role pkg.UserRole) bool {
	for _, r := range u.roles {
		if r == role {
			return true
		}
	}
	return false
}
