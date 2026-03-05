package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/userservice/internal/domain"
	"github.com/nullableocean/grpcservices/userservice/internal/errs"
)

var (
	ErrUsernameAlreadyExist = fmt.Errorf("%w: username already exist", errs.ErrInvalidData)
)

type UserStore struct {
	store     map[string]*domain.User
	usernames map[string]struct{}
	nextId    atomic.Int64

	mu sync.RWMutex
}

func NewUserStore() *UserStore {
	return &UserStore{
		store:     make(map[string]*domain.User),
		usernames: make(map[string]struct{}),
		nextId:    atomic.Int64{},
		mu:        sync.RWMutex{},
	}
}

func (s *UserStore) Save(ctx context.Context, user *domain.User) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ex := s.usernames[user.Username]; ex {
		return nil, fmt.Errorf("%w: username: %s", ErrUsernameAlreadyExist, user.Username)
	}

	if user.UUID == "" {
		return nil, fmt.Errorf("%w: empty uuid", errs.ErrInvalidData)
	}

	s.store[user.UUID] = user
	s.usernames[user.Username] = struct{}{}

	return user, nil
}

func (s *UserStore) Get(ctx context.Context, uuid string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, found := s.store[uuid]
	if !found {
		return nil, fmt.Errorf("%w: user not found", errs.ErrNotFound)
	}

	return u, nil
}

func (s *UserStore) Delete(ctx context.Context, uuid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, uuid)

	return nil
}

func (s *UserStore) Update(ctx context.Context, uuid string, updateInfo *domain.UpdateUserDto) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, found := s.store[uuid]
	if !found {
		return fmt.Errorf("%w: user not found", errs.ErrNotFound)
	}

	if updateInfo.Username != "" {
		if _, ex := s.usernames[updateInfo.Username]; ex {
			return fmt.Errorf("%w: username: %s", ErrUsernameAlreadyExist, updateInfo.Username)
		}

		u.Username = updateInfo.Username
	}

	if updateInfo.PassHash != "" {
		u.PassHash = updateInfo.PassHash
	}

	if updateInfo.Roles != nil {
		u.Roles = updateInfo.Roles
	}

	return nil
}
