package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
)

var (
	ErrUsernameAlreadyExist = fmt.Errorf("%w: username already exist", service.ErrInvalidData)
)

type UserStore struct {
	store     map[int64]*domain.User
	usernames map[string]struct{}
	nextId    atomic.Int64

	mu sync.RWMutex
}

func NewUserStore() *UserStore {
	return &UserStore{
		store:     make(map[int64]*domain.User),
		usernames: make(map[string]struct{}),
		nextId:    atomic.Int64{},
		mu:        sync.RWMutex{},
	}
}

func (s *UserStore) Save(ctx context.Context, userData *domain.CreateUserDto) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ex := s.usernames[userData.Username]; ex {
		return nil, fmt.Errorf("%w: username: %s", ErrUsernameAlreadyExist, userData.Username)
	}

	id := s.nextId.Add(1)
	userData.Id = id

	newUser := domain.NewUser(userData)

	s.store[id] = newUser
	s.usernames[newUser.Username()] = struct{}{}

	return newUser, nil
}

func (s *UserStore) Get(ctx context.Context, id int64) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, found := s.store[id]
	if !found {
		return nil, fmt.Errorf("%w: user not found. id: %d", service.ErrNotFound, id)
	}

	return u, nil
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, found := s.store[id]
	if !found {
		return fmt.Errorf("%w: user not found. id: %d", service.ErrNotFound, id)
	}

	delete(s.store, id)
	delete(s.usernames, u.Username())
	u.Delete()

	return nil
}

func (s *UserStore) Update(ctx context.Context, id int64, updateData *domain.UpdateUserDto) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, found := s.store[id]
	if !found {
		return fmt.Errorf("%w: user not found. id: %d", service.ErrNotFound, id)
	}

	u.SetRoles(updateData.Roles)

	return nil
}
