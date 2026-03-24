package server

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *OrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("order_grpc_server").Start(ctx, "create_order")
	defer span.End()

	userUUID, ok := ctx.Value(shared_inters.UserCtxKey).(string)
	if !ok || userUUID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	span.SetAttributes(attribute.String("user_uuid", userUUID))
	logger := srv.logger.With(zap.String("user_uuid", userUUID))
	logger.Info("grpc received call for create order")

	rolesRaw := ctx.Value(shared_inters.RolesCtxKey)

	var roles []model.UserRole
	if rolesRaw != nil {
		r, ok := rolesRaw.([]string)
		if ok {
			roles = make([]model.UserRole, len(r))
			for i, roleStr := range r {
				roles[i] = model.UserRole(roleStr)
			}
		} else {
			logger.Warn("roles not casted to []string")
		}
	} else {
		logger.Warn("roles dont provided")
	}

	params, err := srv.mapCreateRequestToDto(req, userUUID, roles)
	if err != nil {
		logger.Warn("invalid create order request", zap.Error(err))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	newOrder, err := srv.orderService.CreateOrder(ctx, params)
	if err != nil {
		span.AddEvent("failed order created")
		logger.Error("failed create order", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	span.AddEvent("order created")
	logger.Info("order created", zap.String("order_uuid", newOrder.UUID))

	return srv.mapOrderToResponse(newOrder), nil
}

func (srv *OrderServer) mapCreateRequestToDto(req *orderv1.CreateOrderRequest, userUUID string, roles []model.UserRole) (*dto.CreateOrderParameters, error) {
	orderSide := mapping.MapProtoSideToOrderSide(req.OrderSide)
	orderType := mapping.MapProtoTypeToOrderType(req.OrderType)

	price := mapping.MapProtoMoneyToDecimal(req.Price)
	quantity := mapping.MapProtoDecimalToDecimal(req.Quantity)

	return &dto.CreateOrderParameters{
		User: &model.User{
			UUID:  userUUID,
			Roles: roles,
		},
		MarketUUID: req.MarketUuid,
		Side:       orderSide,
		Type:       orderType,
		Price:      price,
		Quantity:   quantity,
	}, nil
}

func (srv *OrderServer) mapOrderToResponse(o *model.Order) *orderv1.CreateOrderResponse {
	return &orderv1.CreateOrderResponse{
		OrderUuid: o.UUID,
		Status:    mapping.MapOrderStatusToProtoStatus(o.Status),
	}
}
