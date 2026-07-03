package order

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func (s *OrderService) OrdersList(ctx context.Context, user *model.User, filters model.OrderListFilter) (model.OrderList, error) {
	ctx, span := otel.Tracer("order_service").Start(ctx, "orders_list")
	defer span.End()

	logger := s.logger.With(zap.String("user_uuid", user.UUID))

	if err := filters.Validate(); err != nil {
		return model.OrderList{}, err
	}

	list, err := s.orderRepo.List(ctx, user.UUID, filters)
	if err != nil {
		logger.Error("failed get list orders from repo", zap.Error(err))

		return model.OrderList{}, err
	}

	return list, nil
}
