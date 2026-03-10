package handlers

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/stockmarket"
	"go.uber.org/zap"
)

type StockmarketOrderCreatedHandler struct {
	stockmarket  *stockmarket.StockmarketService
	orderService *order.OrderService
	logger       *zap.Logger
}

func NewStockmarketCreatedOrderHandler(logger *zap.Logger, orderService *order.OrderService, stockmarket *stockmarket.StockmarketService) *StockmarketOrderCreatedHandler {
	return &StockmarketOrderCreatedHandler{
		stockmarket:  stockmarket,
		orderService: orderService,
		logger:       logger,
	}
}

func (h *StockmarketOrderCreatedHandler) Handle(ctx context.Context, e inside.Event) {
	event, ok := e.(*inside.OrderCreatedEvent)
	if !ok {
		h.logger.Error("unexpected event type in stockmarket order created events handler",
			zap.String("expected", string(inside.EVENT_CREATED_ORDER)),
			zap.String("got", e.EventType()))
		return
	}
	h.logger.Info("process created order event with stockmarket", zap.String("order_uuid", event.OrderUuid))

	order, err := h.orderService.FindOrder(ctx, event.OrderUuid)
	if err != nil {
		h.logger.Warn("find order by created event error", zap.String("order_uuid", event.OrderUuid), zap.Error(err))
		return
	}

	if err := h.stockmarket.Process(ctx, order); err != nil {
		h.logger.Error("failed process order in stockmarket service",
			zap.String("order_uuid", event.OrderUuid),
			zap.Error(err),
		)
	}
}
