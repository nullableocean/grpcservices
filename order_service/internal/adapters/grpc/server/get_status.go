package server

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *OrderServer) GetOrderStatus(ctx context.Context, req *orderv1.GetStatusRequest) (*orderv1.GetStatusResponse, error) {
	ctx, span := otel.Tracer("order_grpc_server").Start(ctx, "get_order_status")
	defer span.End()

	userUUID, ok := shared_inters.UserUUIDFromContext(ctx)
	if !ok || userUUID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	orderUUID := req.OrderUuid

	span.SetAttributes(attribute.String("user_uuid", userUUID))
	span.SetAttributes(attribute.String("order_uuid", orderUUID))

	logger := srv.logger.With(zap.String("user_uuid", userUUID), zap.String("order_uuid", orderUUID))
	logger.Info("grpc received call for get order status")

	o, err := srv.orderService.GetOrder(ctx, req.OrderUuid, userUUID)
	if err != nil {
		span.AddEvent("failed get order status")
		logger.Warn("failed get order", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	return srv.mapOrderStatusToResponse(o.Status), nil
}

func (srv *OrderServer) mapOrderStatusToResponse(s model.OrderStatus) *orderv1.GetStatusResponse {
	return &orderv1.GetStatusResponse{
		Status: mapping.MapOrderStatusToProtoStatus(s),
	}
}
