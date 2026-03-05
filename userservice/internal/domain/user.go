package domain

import (
	"fmt"

	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/userservice/internal/errs"
)

type User struct {
	UUID     string       `json:"uuid"`
	Username string       `json:"username"`
	Roles    *roles.Roles `json:"user_roles"`
	PassHash string       `json:"-"`
}

type CreateUserDto struct {
	Username string
	Password string
	Roles    []roles.UserRole
}

func (data *CreateUserDto) Validate() error {
	if data.Username == "" {
		return fmt.Errorf("%w: empty username", errs.ErrInvalidData)
	}
	if data.Password == "" {
		return fmt.Errorf("%w: empty password", errs.ErrInvalidData)
	}

	return nil
}

type UpdateUserDto struct {
	Username string
	PassHash string
	Roles    *roles.Roles
}
