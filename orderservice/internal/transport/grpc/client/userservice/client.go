package userservice

import (
	"context"

	userv1 "github.com/nullableocean/grpcservices/api/gen/user/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/mapping"
	"github.com/nullableocean/grpcservices/shared/roles"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserClient struct {
	grpcClient userv1.UserClient

	logger *zap.Logger
}

func NewUserClient(logger *zap.Logger, client userv1.UserClient) *UserClient {
	return &UserClient{
		grpcClient: client,
		logger:     logger,
	}
}

func (client *UserClient) GetUserRoles(ctx context.Context, userUuid string) ([]roles.UserRole, error) {
	ctx, span := otel.Tracer("user_client").Start(ctx, "get_roles_request")
	defer span.End()

	request := &userv1.UserRolesRequest{
		UserUuid: userUuid,
	}

	client.logger.Info("request roles from user grpc server")

	resp, err := client.grpcClient.GetUserRoles(ctx, request)
	if err != nil {
		client.logger.Warn("failed get user roles by grpc", zap.Error(err))

		s, ok := status.FromError(err)
		if ok {
			if s.Code() == codes.NotFound {
				return nil, errs.ErrNotFound
			}
		}

		return nil, err
	}

	return mapping.MapProtoUserRolesResponseToUserRoles(resp), nil
}
