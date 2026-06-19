package order

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func (s *OrderService) UpdateOrder(ctx context.Context, orderUUID string, data *dto.UpdateOrderParameters) error {
	ctx, span := otel.Tracer("order_service").Start(ctx, "update_order")
	defer span.End()

	logger := s.logger.With(zap.String("order_uuid", orderUUID))

	if err := data.Validate(); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return err
	}

	order, err := s.findOrder(ctx, orderUUID)
	if err != nil {
		logger.Error("failed to find order", zap.Error(err))
		return err
	}

	oldStatus := order.Status
	if err := s.applyStatusTransition(order, data.Status); err != nil {
		logger.Error("invalid status update", zap.Error(err))
		s.metrics.OrderFailedUpdate(ctx)

		return err
	}

	event := s.orderFactory.CreateUpdatedEvent(orderUUID, oldStatus, data.Status)

	if err := s.orderRepo.Update(ctx, order, event); err != nil {
		logger.Error("failed to save update", zap.Error(err))
		s.metrics.OrderFailedUpdate(ctx)

		return fmt.Errorf("failed to save updates: %w", errs.ErrCantUpdate)
	}

	s.recordUpdatedMetric(ctx, data.Status)
	logger.Debug("order updated successfully", zap.String("new_status", string(data.Status)))

	return nil
}

func (s *OrderService) applyStatusTransition(order *model.Order, newStatus model.OrderStatus) error {
	if !order.Status.CanTransitTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s: %w", order.Status, newStatus, errs.ErrCantUpdate)
	}

	order.Status = newStatus

	return nil
}

func (s *OrderService) recordUpdatedMetric(ctx context.Context, newStatus model.OrderStatus) {
	switch newStatus {
	case model.OrderStatusCompleted:
		s.metrics.OrderCompleted(ctx)
	case model.OrderStatusRejected:
		s.metrics.OrderRejected(ctx)
	case model.OrderStatusCancelled:
		s.metrics.OrderCancelled(ctx)
	}
}
