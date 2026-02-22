package domain

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/nullableocean/grpcservices/pkg/roles"
)

type User struct {
	id       int64
	username string
	roles    []roles.UserRole
	passHash string

	mu        sync.RWMutex
	deletedAt time.Time
	deleted   int32
}

type CreateUserDto struct {
	Id       int64
	Username string
	PassHash string
	Roles    []roles.UserRole
}

type UpdateUserDto struct {
	Roles []roles.UserRole
}

func NewUser(dto *CreateUserDto) *User {
	return &User{
		id:       dto.Id,
		username: dto.Username,
		passHash: dto.PassHash,
		roles:    dto.Roles,
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
	u.mu.RLock()
	defer u.mu.RUnlock()

	copyOut := make([]roles.UserRole, len(u.roles))
	copy(copyOut, u.roles)

	return copyOut
}

func (u *User) SetRoles(roles []roles.UserRole) {
	u.mu.Lock()
	u.roles = roles
	u.mu.Unlock()
}

func (u *User) AddRole(role roles.UserRole) {
	u.mu.Lock()
	u.roles = append(u.roles, role)
	u.mu.Unlock()
}

func (u *User) DeleteRole(role roles.UserRole) {
	u.mu.Lock()
	var ind int
	for i, r := range u.roles {
		if r == role {
			ind = i
			break
		}
	}

	u.roles = append(u.roles[:ind], u.roles[ind+1:]...)
	u.mu.Unlock()
}

func (u *User) HasRole(role roles.UserRole) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, r := range u.roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) IsDeleted() bool {
	return atomic.LoadInt32(&u.deleted) == 1
}

func (u *User) DeletedAt() time.Time {
	return u.deletedAt
}

func (u *User) Delete() {
	if atomic.CompareAndSwapInt32(&u.deleted, 0, 1) {
		u.deletedAt = time.Now()
	}
}
