package handlers

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/amqp/writer"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type AmqpOrderCreatedHandler struct {
	writer *writer.CreatedEventWriter
	logger *zap.Logger
}

func NewAmqpOrderCreatedHandler(logger *zap.Logger, writer *writer.CreatedEventWriter) *AmqpOrderCreatedHandler {
	return &AmqpOrderCreatedHandler{
		writer: writer,
		logger: logger,
	}
}

func (h *AmqpOrderCreatedHandler) Handle(ctx context.Context, e inside.Event) {
	ctx, span := otel.Tracer("amqp_created_event_handler").Start(ctx, "handle_event")
	defer span.End()

	event, ok := e.(*inside.OrderCreatedEvent)
	if !ok {
		h.logger.Error("unexpected event type in order created events handler",
			zap.String("expected", string(inside.EVENT_CREATED_ORDER)),
			zap.String("got", e.EventType()))
		return
	}

	h.logger.Info("send event to broker", zap.String("order_uuid", event.Order.UUID))
	if err := h.writer.Write(ctx, event); err != nil {
		h.logger.Error("failed to write order created event to Kafka",
			zap.Error(err),
			zap.String("order_uuid", event.Order.UUID))
	}
}
