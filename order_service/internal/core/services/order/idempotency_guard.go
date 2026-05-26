package order

import (
	"context"
	"errors"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

type IdempotencyGuard struct {
	cache  ports.IdempotencyCache
	logger *zap.Logger
}

func NewIdempotencyGuard(cache ports.IdempotencyCache, logger *zap.Logger) *IdempotencyGuard {
	return &IdempotencyGuard{
		cache:  cache,
		logger: logger,
	}
}

func (g *IdempotencyGuard) Reserve(ctx context.Context, key string) (bool, error) {
	ok, err := g.cache.SetIfNotExist(ctx, key, &model.IdempotencyData{
		Status: model.IdempotencyProcessing,
	})

	if err != nil {
		g.logger.Error("failed to set idempotent key in cache", zap.String("key", key), zap.Error(err))
		return false, errs.ErrIdempotencyInternal
	}

	return ok, nil
}

func (g *IdempotencyGuard) GetExisting(ctx context.Context, key string) (*model.IdempotencyData, error) {
	cached, err := g.cache.Get(ctx, key)
	if errors.Is(err, errs.ErrNotFound) || cached == nil {
		return nil, errs.ErrIdempotencyKeyNotFound
	}

	if err != nil {
		g.logger.Error("failed to get idempotency data from cache", zap.String("key", key), zap.Error(err))
		return nil, errs.ErrIdempotencyInternal
	}

	return cached, nil
}

func (g *IdempotencyGuard) SetFailed(ctx context.Context, key string) error {
	return g.update(ctx, key, &model.IdempotencyData{Status: model.IdempotencyFailed})
}

func (g *IdempotencyGuard) SetCompleted(ctx context.Context, key string, orderUUID string) error {
	return g.update(ctx, key, &model.IdempotencyData{
		Status:    model.IdempotencyCompleted,
		OrderUUID: orderUUID,
	})
}

func (g *IdempotencyGuard) SetProcessing(ctx context.Context, key string) error {
	return g.update(ctx, key, &model.IdempotencyData{
		Status: model.IdempotencyProcessing,
	})
}

func (g *IdempotencyGuard) update(ctx context.Context, key string, data *model.IdempotencyData) error {
	err := g.cache.Update(ctx, key, data)

	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrIdempotencyKeyNotFound
		}

		return errs.ErrIdempotencyInternal
	}

	return nil
}
