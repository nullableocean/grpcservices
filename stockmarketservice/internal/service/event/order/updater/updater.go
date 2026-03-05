package updater

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
)

type UpdateWriter interface {
	Write(ctx context.Context, event *domain.OrderUpdate) error
}

type OrderUpdater struct {
	updateWriter UpdateWriter
}

func NewOrderUpdater(updateWriter UpdateWriter) *OrderUpdater {
	return &OrderUpdater{
		updateWriter: updateWriter,
	}
}

func (w *OrderUpdater) Pending(ctx context.Context, orderUuid string) error {
	event := &domain.OrderUpdate{
		UUID:      uuid.NewString(),
		OrderUuid: orderUuid,
		NewStatus: order.ORDER_STATUS_PENDING,
		CreatedAt: time.Now(),
	}

	return w.updateWriter.Write(ctx, event)
}
func (w *OrderUpdater) Reject(ctx context.Context, orderUuid string) error {
	event := &domain.OrderUpdate{
		UUID:      uuid.NewString(),
		OrderUuid: orderUuid,
		NewStatus: order.ORDER_STATUS_REJECTED,
		CreatedAt: time.Now(),
	}

	return w.updateWriter.Write(ctx, event)
}
func (w *OrderUpdater) Complete(ctx context.Context, orderUuid string) error {
	event := &domain.OrderUpdate{
		UUID:      uuid.NewString(),
		OrderUuid: orderUuid,
		NewStatus: order.ORDER_STATUS_COMPLETED,
		CreatedAt: time.Now(),
	}

	return w.updateWriter.Write(ctx, event)
}
