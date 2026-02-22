package user

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/auth"
	"github.com/nullableocean/grpcservices/pkg/roles"
)

type UserStore interface {
	Save(ctx context.Context, userData *domain.CreateUserDto) (*domain.User, error)
	Get(ctx context.Context, id int64) (*domain.User, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, id int64, updateData *domain.UpdateUserDto) error
}

type UserService struct {
	passService *auth.PasswordService
	store       UserStore
}

func NewUserService(store UserStore) *UserService {
	return &UserService{
		passService: &auth.PasswordService{},
		store:       store,
	}
}

func (s *UserService) CreateUser(ctx context.Context, username string, pass string, roles []roles.UserRole) (*domain.User, error) {

	passHash, err := s.passService.GetHashForPassword(pass)
	if err != nil {
		return nil, fmt.Errorf("cant get hash for password: %w", service.ErrInvalidData)
	}

	createDto := &domain.CreateUserDto{
		Id:       0,
		Username: username,
		PassHash: passHash,
		Roles:    roles,
	}

	return s.store.Save(ctx, createDto)
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.store.Get(ctx, id)
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.store.Delete(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, updateData *domain.UpdateUserDto) error {
	return s.store.Update(ctx, id, updateData)
}
