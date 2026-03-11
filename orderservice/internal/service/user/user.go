package user

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/shared/roles"
	"go.opentelemetry.io/otel"
)

type UserClient interface {
	GetUserRoles(ctx context.Context, userUuid string) ([]roles.UserRole, error)
}

type UserService struct {
	client UserClient
}

func NewUserService(client UserClient) *UserService {
	return &UserService{
		client: client,
	}
}

func (s *UserService) GetUser(ctx context.Context, userUuid string) (*domain.User, error) {
	ctx, span := otel.Tracer("user_service").Start(ctx, "get_user")
	defer span.End()

	rls, err := s.client.GetUserRoles(ctx, userUuid)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		UUID:  userUuid,
		Roles: roles.NewRoles(rls...),
	}

	return user, err
}
