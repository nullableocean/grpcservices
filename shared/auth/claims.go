package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims

	UserUUID string   `json:"uuid"`
	Roles    []string `json:"roles"`
}

func (c *UserClaims) Validate() error {
	if c.UserUUID == "" {
		return errors.New("empty user uuid in token")
	}

	return nil
}

func (c *UserClaims) GetUserUUID() string {
	return c.UserUUID
}

func (c *UserClaims) GetRoles() []string {
	rlsCopy := make([]string, len(c.Roles))
	copy(rlsCopy, c.Roles)

	return rlsCopy
}
