package pkg

type UserRole int

const (
	USER_GUEST UserRole = iota + 1
	USER_VERIFIED
	USER_SELLER
	USER_MODER
	USER_ADMIN
)
