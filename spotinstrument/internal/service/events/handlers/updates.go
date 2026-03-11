package handlers

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/eventbus"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/events"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/transport/amqp/writer"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type MarketUpdatesEventHandler struct {
	writer *writer.UpdateWriter
	logger *zap.Logger
}

func NewMarketUpdatesEventHandler(logger *zap.Logger, writer *writer.UpdateWriter) *MarketUpdatesEventHandler {
	return &MarketUpdatesEventHandler{
		writer: writer,
		logger: logger,
	}
}

func (h *MarketUpdatesEventHandler) Handle(ctx context.Context, e eventbus.Event) {
	event, ok := e.(*events.MarketUpdateEvent)
	if !ok {
		h.logger.Warn("unexpected event type in update event handler", zap.String("type", e.EventType()))
		return
	}

	ctx, span := otel.Tracer("markets_update_event_handler").Start(ctx, "handler_market_update_event")
	defer span.End()

	h.logger.Info("write market update event to broker", zap.String("market_uuid", event.MarketUuid))
	if err := h.writer.Write(ctx, event); err != nil {
		h.logger.Error("failed to write market update event to Kafka",
			zap.Error(err),
			zap.String("market_uuid", event.MarketUuid))
	}
}
