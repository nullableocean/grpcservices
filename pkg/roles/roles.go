package roles

type UserRole int

const (
	USER_GUEST UserRole = iota + 1
	USER_VERIFIED
	USER_SELLER
	USER_MODER
	USER_ADMIN
)

func MapInString(r UserRole) string {
	switch r {
	case USER_GUEST:
		return "guest"
	case USER_VERIFIED:
		return "verified"
	case USER_SELLER:
		return "seller"
	case USER_MODER:
		return "moder"
	case USER_ADMIN:
		return "admin"
	}

	return ""
}
