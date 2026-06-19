package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func (s *OrderService) CreateOrder(ctx context.Context, data *dto.CreateOrderParameters) (*model.Order, error) {
	ctx, span := otel.Tracer("order_service").Start(ctx, "create_order")
	defer span.End()

	logger := s.logger.With(
		zap.String("user_uuid", data.User.UUID),
		zap.String("market_uuid", data.MarketUUID),
		zap.String("idempotency_key", data.IdempotencyKey),
	)

	if err := data.Validate(); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return nil, err
	}

	span.SetAttributes(attribute.String("idempotency_key", data.IdempotencyKey))

	ok, err := s.idempotencyGuard.Reserve(ctx, data.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if !ok {
		idemData, err := s.idempotencyGuard.GetExisting(ctx, data.IdempotencyKey)
		if err != nil && !errors.Is(err, errs.ErrIdempotencyKeyNotFound) {
			logger.Error("failed get existing idempotency data", zap.Error(err))
			return nil, err
		}

		if idemData == nil {
			logger.Error("idempotency data not found by key")
			return nil, errs.ErrIdempotencyInternal
		}

		switch {
		case idemData.IsCompleted():
			if idemData.OrderUUID == "" {
				logger.Error("idempotency error. emypty order uuid in cached data")
				return nil, errs.ErrIdempotencyInternal
			}

			order, err := s.orderRepo.FindByUUID(ctx, idemData.OrderUUID)
			if err != nil {
				logger.Error("failed get order from repository", zap.String("order_uuid", idemData.OrderUUID), zap.Error(err))
				return nil, err
			}

			logger.Debug("order found by idempotency cached uuid")
			return order, nil
		case idemData.IsProcessing():
			return nil, errs.ErrIdempotencyProcessing
		case idemData.IsFailed():
			logger.Debug("previous request by idempotency key was failed, go retry", zap.String("previous_error", idemData.LastError))

			err := s.idempotencyGuard.SetProcessing(ctx, data.IdempotencyKey)
			if err != nil {
				logger.Error("failed set processing idem key")
				return nil, err
			}
		}
	}

	if err := s.accessService.CanCreateOrder(ctx, data.User, data); err != nil {
		logger.Debug("access denied", zap.Error(err))

		s.idempotencyGuard.SetFailed(ctx, data.IdempotencyKey)

		return nil, errors.Join(errs.ErrNotAllowed, err)
	}

	if err := s.marketValidator.Validate(ctx, data.MarketUUID, data.User.Roles); err != nil {
		logger.Error("market validation failed", zap.Error(err))
		s.metrics.OrderFailedCreate(ctx)

		s.idempotencyGuard.SetFailed(ctx, data.IdempotencyKey)

		return nil, err
	}

	newOrder := s.orderFactory.CreateOrder(data)
	event := s.orderFactory.CreateCreatedEvent(newOrder)

	if err := s.orderRepo.Save(ctx, newOrder, event); err != nil {
		logger.Error("failed to save order", zap.Error(err))
		s.metrics.OrderFailedCreate(ctx)

		s.idempotencyGuard.SetFailed(ctx, data.IdempotencyKey)

		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	logger.Debug("order created successfully", zap.String("order_uuid", newOrder.UUID))
	s.metrics.OrderCreated(ctx)

	s.idempotencyGuard.SetCompleted(ctx, data.IdempotencyKey, newOrder.UUID)

	return newOrder, nil
}
