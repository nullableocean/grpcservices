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
)

func (srv *OrderServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	ctx, span := otel.Tracer("order_grpc_server").Start(ctx, "get_order")
	defer span.End()

	user, err := srv.extractUserFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not extracted from context")
	}

	span.SetAttributes(attribute.String("user_uuid", user.UUID))
	span.SetAttributes(attribute.String("order_uuid", req.OrderUuid))

	logger := srv.logger.With(zap.String("user_uuid", user.UUID), zap.String("order_uuid", req.OrderUuid))
	logger.Debug("grpc received call for get order")

	o, err := srv.orderService.GetOrder(ctx, req.OrderUuid, user)
	if err != nil {
		span.AddEvent("failed get order")
		logger.Warn("failed get order", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	return srv.createOrderResponse(o), nil
}

func (srv *OrderServer) createOrderResponse(order *model.Order) *orderv1.GetOrderResponse {
	return &orderv1.GetOrderResponse{
		Order: mapping.MapOrderToProtoOrder(order),
	}
}
