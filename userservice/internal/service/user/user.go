package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/userservice/internal/auth"
	"github.com/nullableocean/grpcservices/userservice/internal/domain"
	"github.com/nullableocean/grpcservices/userservice/internal/errs"
)

type UserStore interface {
	Save(ctx context.Context, newUser *domain.User) (*domain.User, error)
	Get(ctx context.Context, uuid string) (*domain.User, error)
	Delete(ctx context.Context, uuid string) error
	Update(ctx context.Context, uuid string, updateData *domain.UpdateUserDto) error
}

type UserService struct {
	passHasher *auth.PasswordHasher
	store      UserStore
}

func NewUserService(store UserStore) *UserService {
	return &UserService{
		passHasher: &auth.PasswordHasher{},
		store:      store,
	}
}

func (s *UserService) GetUser(ctx context.Context, uuid string) (*domain.User, error) {
	return s.store.Get(ctx, uuid)
}

func (s *UserService) CreateUser(ctx context.Context, createData *domain.CreateUserDto) (*domain.User, error) {
	if err := createData.Validate(); err != nil {
		return nil, err
	}

	passHash, err := s.passHasher.GetHashForPassword(createData.Password)
	if err != nil {
		return nil, fmt.Errorf("cant get hash for password: %w", errs.ErrInvalidData)
	}

	rls := createData.Roles
	if len(rls) == 0 {
		rls = s.getBasicRoles()
	}

	newUser := &domain.User{
		UUID:     uuid.NewString(),
		Username: createData.Username,
		Roles:    roles.NewRoles(rls...),
		PassHash: passHash,
	}

	return s.store.Save(ctx, newUser)
}

func (s *UserService) DeleteUser(ctx context.Context, uuid string) error {
	return s.store.Delete(ctx, uuid)
}

func (s *UserService) UpdateUser(ctx context.Context, uuid string, updateData *domain.UpdateUserDto) error {
	return s.store.Update(ctx, uuid, updateData)
}

func (s *UserService) getBasicRoles() []roles.UserRole {
	return []roles.UserRole{roles.USER_GUEST}
}
