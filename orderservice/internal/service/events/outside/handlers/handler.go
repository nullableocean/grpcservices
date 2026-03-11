package handlers

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/outside"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type UpdateEventStore interface {
	Save(ctx context.Context, event *outside.UpdateStatusEvent) error
	Update(ctx context.Context, event *outside.UpdateStatusEvent) error
	Find(ctx context.Context, uuid string) (*outside.UpdateStatusEvent, error)
}

type UpdateEventHandler struct {
	store    UpdateEventStore
	oService *order.OrderService

	logger *zap.Logger
}

func NewUpdateEventHandler(logger *zap.Logger, oService *order.OrderService, store UpdateEventStore) *UpdateEventHandler {
	return &UpdateEventHandler{
		oService: oService,
		store:    store,
		logger:   logger,
	}
}

func (h *UpdateEventHandler) Handle(ctx context.Context, event *outside.UpdateStatusEvent) error {
	ctx, span := otel.Tracer("update_order_event_handler").Start(ctx, "handle_update_event")
	defer span.End()

	span.SetAttributes(attribute.String("event_uuid", event.UUID))
	span.SetAttributes(attribute.String("order_uuid", event.OrderUuid))

	ev, _ := h.store.Find(ctx, event.UUID)
	if ev != nil {
		if ev.ProcessingStatus != outside.EVENT_STATUS_ERROR {
			span.AddEvent("dublicate hit")
			return outside.ErrEventAlreadyHandled
		}

		event = ev
		event.ProcessingStatus = outside.EVENT_STATUS_PROCESSING
		err := h.store.Update(ctx, event)
		if err != nil {
			span.AddEvent("update event status error")
			return err
		}
	} else {
		event.ProcessingStatus = outside.EVENT_STATUS_PROCESSING
		err := h.store.Save(ctx, event)
		if err != nil {
			span.AddEvent("failed store event")
			h.logger.Error("failed store event", zap.Error(err))

			return err
		}
	}

	newOrderStatus, err := h.oService.ChangeStatus(ctx, event.OrderUuid, event.NewStatus)
	if err != nil {
		span.AddEvent("change order status error")
		h.logger.Warn("failed change order status", zap.Error(err))

		event.ProcessingStatus = outside.EVENT_STATUS_ERROR
		if err := h.store.Update(ctx, event); err != nil {
			span.AddEvent("update event status error")
			h.logger.Error("failed update event status", zap.Error(err))
		}

		return err
	}

	h.logger.Info("order status changed",
		zap.String("order_uuid", event.OrderUuid),
		zap.String("new_status", newOrderStatus.String()),
	)

	event.ProcessingStatus = outside.EVENT_STATUS_PROCESSED
	if err := h.store.Update(ctx, event); err != nil {
		span.AddEvent("update event status error")
		h.logger.Error("failed update event status", zap.Error(err))
	}

	return nil
}
