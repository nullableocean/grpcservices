package roles

import (
	"slices"
)

type UserRole int

const (
	USER_GUEST UserRole = iota + 1
	USER_VERIFIED
	USER_SELLER
	USER_MODER
	USER_ADMIN
)

type Roles struct {
	roles map[UserRole]struct{}
}

func NewRoles(roles ...UserRole) *Roles {
	m := make(map[UserRole]struct{}, len(roles))

	for _, r := range roles {
		m[r] = struct{}{}
	}

	return &Roles{
		roles: m,
	}
}

func (r *Roles) Has(role UserRole) bool {
	_, ex := r.roles[role]
	return ex
}

func (r *Roles) Add(role UserRole) {
	if r.roles == nil {
		r.roles = make(map[UserRole]struct{})
	}

	r.roles[role] = struct{}{}
}

func (r *Roles) Remove(role UserRole) {
	delete(r.roles, role)
}

func (r *Roles) GetSlice() []UserRole {
	out := make([]UserRole, 0, len(r.roles))

	for r := range r.roles {
		out = append(out, r)
	}

	return out
}

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

func MapSliceToStrings(r []UserRole) []string {
	out := make([]string, 0, len(r))

	for _, role := range r {
		out = append(out, MapInString(role))
	}

	return out
}

// desc sort admin -> moder -> ...
func SortRolesDesc(rls []UserRole) {
	slices.SortFunc(rls, func(a, b UserRole) int {
		if a > b {
			return -1
		}
		return 1
	})
}
