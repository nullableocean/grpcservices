package server

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *OrderServer) GetOrderStatus(ctx context.Context, req *orderv1.GetStatusRequest) (*orderv1.GetStatusResponse, error) {
	ctx, span := otel.Tracer("order_grpc_server").Start(ctx, "get_order_status")
	defer span.End()

	user, err := srv.extractUserFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not extracted from context")
	}

	span.SetAttributes(attribute.String("user_uuid", user.UUID))
	span.SetAttributes(attribute.String("order_uuid", req.OrderUuid))

	logger := srv.logger.With(zap.String("user_uuid", user.UUID), zap.String("order_uuid", req.OrderUuid))
	logger.Debug(ctx, "grpc received call for get order status")

	o, err := srv.orderService.GetOrder(ctx, req.OrderUuid, user)
	if err != nil {
		span.AddEvent("failed get order status")
		logger.Warn(ctx, "failed get order", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	return srv.createOrderStatusResponse(o), nil
}

func (srv *OrderServer) createOrderStatusResponse(order *model.Order) *orderv1.GetStatusResponse {
	return &orderv1.GetStatusResponse{
		Status:    mapping.MapOrderStatusToProtoStatus(order.Status),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}
}
