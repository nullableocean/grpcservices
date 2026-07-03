package model

import (
	"encoding/json"
	"fmt"
)

type UserRole string

const (
	UserRoleGuest       UserRole = "GUEST"
	UserRoleTrader      UserRole = "TRADER"
	UserRoleMarketMaker UserRole = "MARKET_MAKER"
	UserRoleModer       UserRole = "MODER"
	UserRoleAdmin       UserRole = "ADMIN"
)

var rolePriority = map[UserRole]int{
	UserRoleGuest:       1,
	UserRoleTrader:      2,
	UserRoleMarketMaker: 3,
	UserRoleModer:       4,
	UserRoleAdmin:       5,
}

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleGuest, UserRoleTrader, UserRoleMarketMaker, UserRoleModer, UserRoleAdmin:
		return true
	}

	return false
}

func (r UserRole) GetPriority() int {
	return rolePriority[r]
}

func (r UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

func (r *UserRole) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	role := UserRole(str)
	if !role.IsValid() {
		return fmt.Errorf("invalid user role: %s", str)
	}

	*r = role
	return nil
}

type User struct {
	UUID  string     `json:"uuid"`
	Roles []UserRole `json:"roles"`
}

func NewUser(uuid string, roles []UserRole) *User {
	return &User{
		UUID:  uuid,
		Roles: roles,
	}
}

func (u *User) HighestPriorityRole() UserRole {
	priorityRole := UserRoleGuest
	weight := 0

	for _, role := range u.Roles {
		if w, ok := rolePriority[role]; ok && w > weight {
			weight = w
			priorityRole = role
		}
	}

	return priorityRole
}
