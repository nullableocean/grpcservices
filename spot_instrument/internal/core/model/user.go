package model

type UserRole string

const (
	UserRoleGuest       UserRole = "GUEST"
	UserRoleTrader      UserRole = "TRADER"
	UserRoleMarketMaker UserRole = "MARKET_MAKER"
	UserRoleModer       UserRole = "MODER"
	UserRoleAdmin       UserRole = "ADMIN"
)

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
