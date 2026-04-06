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
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func (s *OrderService) CreateOrder(ctx context.Context, data *dto.CreateOrderParameters) (*model.Order, error) {
	ctx, span := otel.Tracer("order_service").Start(ctx, "create_order")
	defer span.End()

	user := data.User
	logger := s.logger.With(
		zap.String("user_uuid", user.UUID),
		zap.String("market_uuid", data.MarketUUID),
		zap.String("idempotency_key", data.IdempotencyKey),
	)

	if err := data.Validate(); err != nil {
		logger.Warn("failed create order. validation error", zap.Error(err))

		return nil, err
	}

	span.SetAttributes(attribute.String("idempotency_key", data.IdempotencyKey))
	ok, err := s.idemCache.SetIfNotExist(ctx, data.IdempotencyKey, &model.IdempotencyData{
		Status: model.IdempotencyProcessing,
	})
	if err != nil {
		logger.Error("failed set idempotent key in cache", zap.Error(err))
		return nil, errs.ErrIdempotencyInternal
	}

	if !ok {
		cached, err := s.idemCache.Get(ctx, data.IdempotencyKey)
		if err != nil {
			logger.Error("failed get idempotency data from cache", zap.Error(err))

			return nil, errs.ErrIdempotencyInternal
		}
		if cached != nil {
			switch cached.Status {
			case model.IdempotencyCompleted:
				if cached.OrderUUID == "" {
					logger.Error("idempotency error. emypty order uuid in cached data")
					return nil, errs.ErrIdempotencyInternal
				}

				order, err := s.orderRepo.FindByUUID(ctx, cached.OrderUUID)
				if err != nil {
					logger.Error("failed get order from repository", zap.String("order_uuid", cached.OrderUUID), zap.Error(err))
					return nil, err
				}

				return order, nil
			case model.IdempotencyProcessing:
				return nil, errs.ErrIdempotencyProcessing
			case model.IdempotencyFailed:
				err := s.idemCache.Delete(ctx, data.IdempotencyKey)
				if err != nil {
					logger.Error("failed delete idempotent key from cache")
					return nil, errs.ErrIdempotencyInternal
				}
			}
		}
	}

	err = s.accessService.CanCreateOrder(ctx, user, data)
	if err != nil {
		logger.Info("failed access")

		s.idemCache.Update(ctx, data.IdempotencyKey, &model.IdempotencyData{
			Status: model.IdempotencyFailed,
		})

		return nil, errors.Join(errs.ErrNotAllowed, err)
	}

	logger.Info("create order")

	_, err = s.spotInstrument.FindMarket(ctx, data.MarketUUID, user.Roles)
	if err != nil {
		s.metrics.OrderFailedCreate(ctx)
		logger.Error("failed create order. failed get market", zap.Error(err))

		cacheErr := s.idemCache.Update(ctx, data.IdempotencyKey, &model.IdempotencyData{
			Status: model.IdempotencyFailed,
		})
		if cacheErr != nil {
			logger.Error("failed update idempotency cache data", zap.Error(cacheErr))
		}

		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("failed create order: failed get market: %w", errs.ErrNotFound)
		}

		if errors.Is(err, ports.ErrNotAllowed) {
			return nil, fmt.Errorf("failed create order: market not allowed for user: %w", errs.ErrNotAllowed)
		}

		return nil, fmt.Errorf("failed create order: failed get market: %w", err)
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
		span.AddEvent("failed save order")

		cacheErr := s.idemCache.Update(ctx, data.IdempotencyKey, &model.IdempotencyData{
			Status: model.IdempotencyFailed,
		})
		if cacheErr != nil {
			logger.Error("failed update idempotency cache data", zap.Error(cacheErr))
		}

		return nil, fmt.Errorf("failed create order: failed save: %w", err)
	}

	s.metrics.OrderCreated(ctx)
	logger.Info("success created", zap.String("order_uuid", newOrder.UUID))

	cacheErr := s.idemCache.Update(ctx, data.IdempotencyKey, &model.IdempotencyData{
		Status:    model.IdempotencyCompleted,
		OrderUUID: newOrder.UUID,
	})
	if cacheErr != nil {
		logger.Error("failed update idempotency cache data", zap.Error(cacheErr))
	}

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
