package model

type UserRole string

const (
	UserRoleGuest       UserRole = "GUEST"
	UserRoleTrader      UserRole = "TRADER"
	UserRoleMarketMaker UserRole = "MARKET_MAKER"
	UserRoleModer       UserRole = "MODER"
	UserRoleAdmin       UserRole = "ADMIN"
)
