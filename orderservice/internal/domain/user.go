package domain

import (
	"github.com/nullableocean/grpcservices/shared/roles"
)

type User struct {
	UUID  string
	Roles *roles.Roles
}

func (u *User) GetUuid() string {
	return u.UUID
}

func (u *User) HasRole(role roles.UserRole) bool {
	return u.Roles.Has(role)
}

func (u *User) GetRoles() []roles.UserRole {
	return u.Roles.GetSlice()
}
