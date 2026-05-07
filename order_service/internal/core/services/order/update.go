package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
		logger.Warn("failed update order. validation error", zap.Error(err))

		return err
	}

	o, err := s.findOrder(ctx, orderUUID)
	if err != nil {
		logger.Error("failed update order. find order error", zap.Error(err))

		return err
	}

	oldStatus := o.Status

	err = s.updateOrderByParams(o, data)
	if err != nil {
		s.metrics.OrderFailedUpdate(ctx)
		logger.Error("failed update order", zap.Error(err))

		return err
	}

	event := &model.EventOrderUpdated{
		UUID:      uuid.NewString(),
		OrderUUID: orderUUID,
		Data: &model.EventUpdatedData{
			NewStatus: &data.Status,
			OldStatus: &oldStatus,
			UpdatedAt: time.Now(),
		},
	}

	err = s.orderRepo.Update(ctx, o, event)
	if err != nil {
		s.metrics.OrderFailedUpdate(ctx)
		return fmt.Errorf("failed save updates: %w", errs.ErrCantUpdate)
	}

	s.recordUpdatedMetric(ctx, data.Status)

	return nil
}

func (s *OrderService) updateOrderByParams(updatingOrder *model.Order, data *dto.UpdateOrderParameters) error {
	if !updatingOrder.Status.CanTransitTo(data.Status) {
		return fmt.Errorf("failed update status from %s to %s: %w", updatingOrder.Status, data.Status, errs.ErrCantUpdate)
	}

	updatingOrder.Status = data.Status

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
