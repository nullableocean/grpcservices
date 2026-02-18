package service

import (
	"fmt"
	"main/order/domain"
	"main/pkg/roles"
	"sync"
	"sync/atomic"
)

var (
	ErrUsernameAlreadyExist = fmt.Errorf("%w: username already exist", ErrInvalidData)
)

type UserService struct {
	store     map[int64]*domain.User
	usernames map[string]struct{}

	nextId atomic.Int64

	mu sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		store:     make(map[int64]*domain.User),
		usernames: make(map[string]struct{}),
		mu:        sync.RWMutex{},
	}
}

func (s *UserService) NewUser(username string, roles []roles.UserRole) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ex := s.usernames[username]; ex {
		return nil, fmt.Errorf("%w: username: %s", ErrUsernameAlreadyExist, username)
	}

	id := s.nextId.Add(1)

	newUser := domain.NewUser(id)
	newUser.SetUsername(username)
	newUser.SetRoles(roles)

	s.store[id] = newUser
	s.usernames[username] = struct{}{}

	return newUser, nil
}

func (s *UserService) GetUser(id int64) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, found := s.store[id]
	if !found {
		return nil, fmt.Errorf("%w: user not found. id: %d", ErrNotFound, id)
	}

	return u, nil
}
