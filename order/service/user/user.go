package user

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/auth"
	"github.com/nullableocean/grpcservices/pkg/roles"
)

var (
	ErrUsernameAlreadyExist = fmt.Errorf("%w: username already exist", service.ErrInvalidData)
)

type UserService struct {
	passService *auth.PasswordService

	store     map[int64]*domain.User
	usernames map[string]struct{}
	nextId    atomic.Int64

	mu sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		passService: &auth.PasswordService{},

		store:     make(map[int64]*domain.User),
		usernames: make(map[string]struct{}),
		mu:        sync.RWMutex{},
	}
}

func (s *UserService) CreateUser(username string, pass string, roles []roles.UserRole) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ex := s.usernames[username]; ex {
		return nil, fmt.Errorf("%w: username: %s", ErrUsernameAlreadyExist, username)
	}

	passHash, err := s.passService.GetHashForPassword(pass)
	if err != nil {
		return nil, fmt.Errorf("cant get hash for password: %w", service.ErrInvalidData)
	}

	id := s.nextId.Add(1)

	newUser := domain.NewUser(id, username, passHash)
	newUser.SetRoles(roles)

	s.store[id] = newUser
	s.usernames[username] = struct{}{}

	return newUser, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, found := s.store[id]
	if !found {
		return nil, fmt.Errorf("%w: user not found. id: %d", service.ErrNotFound, id)
	}

	return u, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, found := s.store[id]
	if !found {
		return fmt.Errorf("%w: user not found. id: %d", service.ErrNotFound, id)
	}

	u.Delete()
	delete(s.store, id)

	return nil
}
