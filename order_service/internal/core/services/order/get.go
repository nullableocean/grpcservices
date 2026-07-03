package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func (s *OrderService) GetOrder(ctx context.Context, orderUUID string, user *model.User) (*model.Order, error) {
	ctx, span := otel.Tracer("order_service").Start(ctx, "get_order")
	defer span.End()

	logger := s.logger.With(zap.String("order_uuid", orderUUID), zap.String("user_uuid", user.UUID))

	o, err := s.findOrder(ctx, orderUUID)
	if err != nil {
		logger.Error("failed get order from repository", zap.Error(err))

		return nil, err
	}

	if err := s.accessService.CanSeeOrder(ctx, user, o); err != nil {
		logger.Error("failed get order by user", zap.String("order_user_uuid", o.UserUUID), zap.Error(err))
		return nil, err
	}

	return o, nil
}

func (s *OrderService) findOrder(ctx context.Context, orderUUID string) (*model.Order, error) {
	o, err := s.orderRepo.FindByUUID(ctx, orderUUID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("find order by uuid error: %w", errs.ErrNotFound)
		}

		return nil, fmt.Errorf("failed find order: %w", errs.ErrNotFound)
	}

	return o, nil
}
