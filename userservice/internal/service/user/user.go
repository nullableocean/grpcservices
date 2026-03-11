package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/userservice/internal/domain"
	"github.com/nullableocean/grpcservices/userservice/internal/errs"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type UserStore interface {
	Save(ctx context.Context, newUser *domain.User) (*domain.User, error)
	Get(ctx context.Context, uuid string) (*domain.User, error)
	Delete(ctx context.Context, uuid string) error
	Update(ctx context.Context, uuid string, updateData *domain.UpdateUserDto) error
}

type PasswordHasher interface {
	Compare(pass, hash string) bool
	GetHash(pass string) (string, error)
}

type UserService struct {
	passHasher PasswordHasher
	store      UserStore

	logger *zap.Logger
}

func NewUserService(logger *zap.Logger, store UserStore, hasher PasswordHasher) *UserService {
	return &UserService{
		passHasher: hasher,
		store:      store,
		logger:     logger,
	}
}

func (s *UserService) GetUser(ctx context.Context, uuid string) (*domain.User, error) {
	ctx, span := otel.Tracer("user_service").Start(ctx, "get_user_roles")
	defer span.End()

	s.logger.Info("get user from store")

	user, err := s.store.Get(ctx, uuid)
	if err != nil {
		span.AddEvent("failed get user")
		s.logger.Error("failed get user from store", zap.Error(err))

		return nil, err
	}

	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, createData *domain.CreateUserDto) (*domain.User, error) {
	ctx, span := otel.Tracer("user_service").Start(ctx, "create_user")
	defer span.End()

	s.logger.Info("creating user")

	if err := createData.Validate(); err != nil {
		span.AddEvent("failed validation")
		s.logger.Warn("validation error", zap.Error(err))

		return nil, err
	}

	passHash, err := s.passHasher.GetHash(createData.Password)
	if err != nil {
		span.AddEvent("failed hashing pass")
		s.logger.Error("failed hash password")

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

	s.logger.Info("store user")
	user, err := s.store.Save(ctx, newUser)
	if err != nil {
		span.AddEvent("failed store user")
		s.logger.Error("failed store user", zap.Error(err))

		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, uuid string) error {
	ctx, span := otel.Tracer("user_service").Start(ctx, "delete_user")
	defer span.End()

	s.logger.Info("delete user", zap.String("user_uuid", uuid))

	return s.store.Delete(ctx, uuid)
}

func (s *UserService) UpdateUser(ctx context.Context, uuid string, updateData *domain.UpdateUserDto) error {
	ctx, span := otel.Tracer("user_service").Start(ctx, "update_user")
	defer span.End()

	s.logger.Info("delete user", zap.String("user_uuid", uuid))

	err := s.store.Update(ctx, uuid, updateData)
	if err != nil {
		span.AddEvent("failed update user")
		s.logger.Error("failed update user", zap.Error(err))
		return err
	}

	return nil
}

func (s *UserService) getBasicRoles() []roles.UserRole {
	return []roles.UserRole{roles.USER_GUEST}
}
