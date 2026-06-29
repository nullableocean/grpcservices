package order

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type idempotencyStatus int

const (
	idempotencyReserved idempotencyStatus = iota
	idempotencyResultExisting
)

type idempotencyCheckResult struct {
	Status idempotencyStatus
	Order  *model.Order
}

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

	result, err := s.checkIdempotency(ctx, logger, data.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if result.Status == idempotencyResultExisting {
		span.SetAttributes(attribute.String("order_uuid", result.Order.UUID))
		return result.Order, nil
	}

	order, err := s.processOrderCreation(ctx, logger, data)
	if err != nil {
		s.markIdempotencyFailed(ctx, logger, data.IdempotencyKey)
		return nil, err
	}

	s.markIdempotencyCompleted(ctx, logger, data.IdempotencyKey, order.UUID)

	span.SetAttributes(attribute.String("order_uuid", order.UUID))
	return order, nil
}

func (s *OrderService) checkIdempotency(ctx context.Context, logger *zap.Logger, idempotencyKey string) (*idempotencyCheckResult, error) {
	ok, err := s.idempotencyGuard.Reserve(ctx, idempotencyKey)
	if err != nil {
		return nil, err
	}

	if ok {
		return &idempotencyCheckResult{Status: idempotencyReserved}, nil
	}

	idemData, err := s.idempotencyGuard.GetExisting(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, errs.ErrIdempotencyKeyNotFound) {
		logger.Error("failed to get existing idempotency data", zap.Error(err))
		return nil, err
	}

	if idemData == nil {
		logger.Error("idempotency data not found by key")
		return nil, errs.ErrIdempotencyInternal
	}

	switch {
	case idemData.IsCompleted():
		return s.handleIdempotencyCompleted(ctx, logger, idemData)
	case idemData.IsProcessing():
		return nil, errs.ErrIdempotencyProcessing
	case idemData.IsFailed():
		return s.handleIdempotencyRetry(ctx, logger, idempotencyKey, idemData)
	default:
		logger.Error("unknown idempotency state")
		return nil, errs.ErrIdempotencyInternal
	}
}

func (s *OrderService) handleIdempotencyCompleted(ctx context.Context, logger *zap.Logger, idemData *model.IdempotencyData) (*idempotencyCheckResult, error) {
	if idemData.OrderUUID == "" {
		logger.Error("idempotency error: empty order uuid in cached data")
		return nil, errs.ErrIdempotencyInternal
	}

	order, err := s.orderRepo.FindByUUID(ctx, idemData.OrderUUID)
	if err != nil {
		logger.Error("failed to get order from repository", zap.String("order_uuid", idemData.OrderUUID), zap.Error(err))
		return nil, err
	}

	logger.Debug("order found by idempotency cached uuid")
	return &idempotencyCheckResult{Status: idempotencyResultExisting, Order: order}, nil
}

func (s *OrderService) handleIdempotencyRetry(ctx context.Context, logger *zap.Logger, idempotencyKey string, idemData *model.IdempotencyData) (*idempotencyCheckResult, error) {
	logger.Debug("previous request by idempotency key was failed, retrying", zap.String("previous_error", idemData.LastError))

	if err := s.idempotencyGuard.SetProcessing(ctx, idempotencyKey); err != nil {
		logger.Error("failed to set processing idempotency key", zap.Error(err))
		return nil, err
	}

	return &idempotencyCheckResult{Status: idempotencyReserved}, nil
}

func (s *OrderService) processOrderCreation(ctx context.Context, logger *zap.Logger, data *dto.CreateOrderParameters) (*model.Order, error) {
	if err := s.authorize(ctx, logger, data); err != nil {
		return nil, err
	}

	if err := s.checkAndIncRateLimits(ctx, logger, data); err != nil {
		return nil, err
	}

	var err error
	defer func() {
		if err != nil {
			s.rollbackRateLimit(ctx, s.logger, data)
		}
	}()

	if err = s.validateMarket(ctx, logger, data); err != nil {
		return nil, err
	}

	createdOrder, err := s.createAndSaveOrder(ctx, logger, data)

	return createdOrder, err
}

func (s *OrderService) authorize(ctx context.Context, logger *zap.Logger, data *dto.CreateOrderParameters) error {
	if err := s.accessService.CanCreateOrder(ctx, data.User, data); err != nil {
		logger.Debug("access denied", zap.Error(err))
		return errors.Join(errs.ErrNotAllowed, err)
	}

	return nil
}

func (s *OrderService) checkAndIncRateLimits(ctx context.Context, logger *zap.Logger, data *dto.CreateOrderParameters) error {
	if err := s.rateLimiter.Check(ctx, *data.User); err != nil {
		logger.Warn("orders rate limit exceeded", zap.Error(err))
		s.metrics.OrderFailedCreate(ctx)

		return err
	}

	return nil
}

func (s *OrderService) rollbackRateLimit(ctx context.Context, logger *zap.Logger, data *dto.CreateOrderParameters) error {
	if err := s.rateLimiter.Rollback(ctx, *data.User); err != nil {
		logger.Warn("orders rate limit failed rollback", zap.Error(err))

		return err
	}

	return nil
}

func (s *OrderService) validateMarket(ctx context.Context, logger *zap.Logger, data *dto.CreateOrderParameters) error {
	if err := s.marketValidator.Validate(ctx, data.MarketUUID, data.User.Roles); err != nil {
		logger.Error("market validation failed", zap.Error(err))
		s.metrics.OrderFailedCreate(ctx)

		return fmt.Errorf("market validation: %w", err)
	}

	return nil
}

func (s *OrderService) createAndSaveOrder(ctx context.Context, logger *zap.Logger, data *dto.CreateOrderParameters) (*model.Order, error) {
	order := s.orderFactory.CreateOrder(data)
	event := s.orderFactory.CreateCreatedEvent(order)

	if err := s.orderRepo.Save(ctx, order, event); err != nil {
		logger.Error("failed to save order", zap.Error(err))
		s.metrics.OrderFailedCreate(ctx)

		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	logger.Debug("order created successfully", zap.String("order_uuid", order.UUID))
	s.metrics.OrderCreated(ctx)

	return order, nil
}

func (s *OrderService) markIdempotencyFailed(ctx context.Context, logger *zap.Logger, idempotencyKey string) {
	if err := s.idempotencyGuard.SetFailed(ctx, idempotencyKey); err != nil {
		logger.Error("failed to mark idempotency key as failed", zap.String("idempotency_key", idempotencyKey), zap.Error(err))
	}
}

func (s *OrderService) markIdempotencyCompleted(ctx context.Context, logger *zap.Logger, idempotencyKey, orderUUID string) {
	if err := s.idempotencyGuard.SetCompleted(ctx, idempotencyKey, orderUUID); err != nil {
		logger.Error("failed to mark idempotency key as completed",
			zap.String("idempotency_key", idempotencyKey),
			zap.String("order_uuid", orderUUID),
			zap.Error(err),
		)
	}
}
