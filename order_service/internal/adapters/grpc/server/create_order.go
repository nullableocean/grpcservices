package server

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *OrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("order_grpc_server").Start(ctx, "create_order")
	defer span.End()

	user, err := srv.extractUserFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not extracted from context")
	}

	span.SetAttributes(attribute.String("user_uuid", user.UUID))
	logger := srv.logger.With(zap.String("user_uuid", user.UUID))
	logger.Debug(ctx, "grpc received call for create order")

	params, err := srv.mapCreateRequestToDto(req, user)
	if err != nil {
		span.AddEvent("failed order created")
		logger.Error(ctx, "failed map request to DTO", zap.Error(err))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	newOrder, err := srv.orderService.CreateOrder(ctx, params)
	if err != nil {
		span.AddEvent("failed order created")
		logger.Error(ctx, "failed create order", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	span.AddEvent("order created")
	logger.Debug(ctx, "order created", zap.String("order_uuid", newOrder.UUID))

	return srv.mapOrderToResponse(newOrder), nil
}

func (srv *OrderServer) mapCreateRequestToDto(req *orderv1.CreateOrderRequest, user *model.User) (*dto.CreateOrderParameters, error) {
	orderSide := mapping.MapProtoSideToOrderSide(req.OrderSide)
	orderType := mapping.MapProtoTypeToOrderType(req.OrderType)

	price := mapping.MapProtoMoneyToDecimal(req.Price)
	quantity := mapping.MapProtoDecimalToDecimal(req.Quantity)

	return &dto.CreateOrderParameters{
		IdempotencyKey: req.IdempotencyKey,
		User:           user,
		MarketUUID:     req.MarketUuid,
		Side:           orderSide,
		Type:           orderType,
		Price:          price,
		Quantity:       quantity,
	}, nil
}

func (srv *OrderServer) mapOrderToResponse(o *model.Order) *orderv1.CreateOrderResponse {
	return &orderv1.CreateOrderResponse{
		Order: mapping.MapOrderToProtoOrder(o),
	}
}
