package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func (s *OrderService) CreateOrder(ctx context.Context, data *dto.CreateOrderParameters) (*model.Order, error) {
	ctx, span := otel.Tracer("order_service").Start(ctx, "create_order")
	defer span.End()

	user := data.User
	logger := s.logger.With(zap.String("user_uuid", user.UUID), zap.String("market_uuid", data.MarketUUID))

	if err := data.Validate(); err != nil {
		logger.Warn("failed create order. validation error", zap.Error(err))

		return nil, err
	}

	err := s.accessService.CanCreateOrder(ctx, user, data)
	if err != nil {
		return nil, errors.Join(errs.ErrNotAllowed, err)
	}

	logger.Info("create order")
	markets, err := s.spotInstrument.ViewMarkets(ctx, user.Roles)
	if err != nil {
		s.metrics.OrderFailedCreate(ctx)
		logger.Error("failed create order. failed get markets", zap.Error(err))

		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("failed create order: failed get markets: %w", errs.ErrNotFound)
		}

		return nil, fmt.Errorf("failed create order: failed get markets: %w", err)
	}

	hasOrderMarket := false
CHECK_MARKET:
	for _, m := range markets {
		if data.MarketUUID == m.UUID {
			hasOrderMarket = true
			break CHECK_MARKET
		}
	}

	if !hasOrderMarket {
		s.metrics.OrderFailedCreate(ctx)
		logger.Error("failed create order. failed find order market in allowed markets", zap.Error(err))

		return nil, fmt.Errorf("failed create order: market uuid not exist in allowed markets: %w", errs.ErrNotAllowed)
	}

	newOrder := s.createOrder(data)
	event := &model.EventOrderCreated{
		UUID:      uuid.NewString(),
		OrderUUID: newOrder.UUID,
		Data: &model.EventCreatedData{
			Order: newOrder,
		},
	}

	err = s.orderRepo.Save(ctx, newOrder, event)
	if err != nil {
		s.metrics.OrderFailedCreate(ctx)
		logger.Error("failed create order. failed save new order", zap.Error(err))

		return nil, fmt.Errorf("failed create order: failed save: %w", err)
	}

	s.metrics.OrderCreated(ctx)
	logger.Info("success created", zap.String("order_uuid", newOrder.UUID))

	return newOrder, nil
}

func (s *OrderService) createOrder(data *dto.CreateOrderParameters) *model.Order {
	now := time.Now()
	return &model.Order{
		UUID:       uuid.NewString(),
		UserUUID:   data.User.UUID,
		MarketUUID: data.MarketUUID,
		Status:     model.OrderStatusCreated,
		Type:       data.Type,
		Side:       data.Side,
		Price:      data.Price,
		Quantity:   data.Quantity,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
