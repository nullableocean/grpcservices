package transport

import (
	"context"
	"errors"

	userv1 "github.com/nullableocean/grpcservices/api/gen/user/v1"
	"github.com/nullableocean/grpcservices/userservice/internal/errs"
	"github.com/nullableocean/grpcservices/userservice/internal/service/user"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServer

	userService *user.UserService
	logger      *zap.Logger
}

func NewUserServer(l *zap.Logger, us *user.UserService) *UserServer {
	return &UserServer{
		userService: us,
		logger:      l,
	}
}

func (s *UserServer) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	ctx, span := otel.Tracer("user_server").Start(ctx, "create_user_request")
	defer span.End()

	dto := MapProtoCreateRequestToCreateDto(req)

	s.logger.Info("create user request", zap.String("username", dto.Username))

	user, err := s.userService.CreateUser(ctx, dto)
	if err != nil {
		span.AddEvent("failed create user")
		s.logger.Info("failed user create", zap.Error(err))

		return nil, s.handleError(err)
	}

	s.logger.Info("user created", zap.String("uuid", user.UUID))

	return MapUserToCreateResponse(user), nil
}

func (s *UserServer) GetUserRoles(ctx context.Context, req *userv1.UserRolesRequest) (*userv1.UserRolesResponse, error) {
	ctx, span := otel.Tracer("user_server").Start(ctx, "get_user_roles_request")
	defer span.End()

	userUuid := req.UserUuid

	s.logger.Info("get user roles request", zap.String("uuid", userUuid))

	user, err := s.userService.GetUser(ctx, userUuid)
	if err != nil {
		span.AddEvent("failed get roles")
		s.logger.Info("failed get user roles request", zap.Error(err))

		return nil, s.handleError(err)
	}

	resp := &userv1.UserRolesResponse{
		UserRoles: MapUserRolesToProtoRoles(user.Roles.GetSlice()),
	}

	return resp, nil
}

func (s *UserServer) handleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, errs.ErrInvalidData) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, errs.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
