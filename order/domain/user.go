package domain

import "github.com/nullableocean/grpcservices/pkg/roles"

type User struct {
	id       int64
	username string
	roles    []roles.UserRole
	passHash string
}

func NewUser(id int64, username string, passHash string) *User {
	return &User{
		id:       id,
		username: username,
		passHash: passHash,
	}
}

func (u *User) Id() int64 {
	return u.id
}

func (u *User) Username() string {
	return u.username
}

func (u *User) PassHash() string {
	return u.passHash
}

func (u *User) Roles() []roles.UserRole {
	return u.roles
}

func (u *User) SetRoles(roles []roles.UserRole) {
	u.roles = roles
}

func (u *User) AddRole(role roles.UserRole) {
	u.roles = append(u.roles, role)
}

func (u *User) HasRole(role roles.UserRole) bool {
	for _, r := range u.roles {
		if r == role {
			return true
		}
	}
	return false
}
