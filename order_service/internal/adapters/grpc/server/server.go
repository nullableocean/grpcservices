package server

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServer struct {
	orderv1.UnimplementedOrderServer

	orderService   order.Service
	updateNotifier ports.UpdateNotifier

	logger *zap.Logger
}

func NewOrderServer(l *zap.Logger, orderService order.Service, updateNotifier ports.UpdateNotifier) *OrderServer {
	return &OrderServer{
		orderService:   orderService,
		updateNotifier: updateNotifier,
		logger:         l,
	}
}

func (srv *OrderServer) getGrpcError(e error) error {
	return mapping.MapErrorToGrpcStatusError(e)
}

func (srv *OrderServer) extractUserFromCtx(ctx context.Context) (*model.User, error) {
	userUUID, ok := shared_inters.UserUUIDFromContext(ctx)
	if !ok || userUUID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	ctxRoles, ok := shared_inters.RolesFromContext(ctx)
	if !ok {
		srv.logger.Warn("roles not provided in context")
	}

	var roles []model.UserRole
	if len(ctxRoles) > 0 {
		roles = make([]model.UserRole, len(ctxRoles))
		for i, roleStr := range ctxRoles {
			roles[i] = model.UserRole(roleStr)
		}
	}

	user := model.NewUser(userUUID, roles)
	return user, nil
}
