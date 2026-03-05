package spot

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/cache"
	"github.com/nullableocean/grpcservices/shared/roles"
	"go.uber.org/zap"
)

type MarketCache interface {
	Get(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error)
	Set(ctx context.Context, roles []roles.UserRole, markets []*domain.Market) error
}

type CachedSpotInstrument struct {
	marketsCache MarketCache

	baseInstrument *SpotInstrument
	logger         *zap.Logger
}

func NewCachedSpotInstrument(baseInstrument *SpotInstrument, cache MarketCache, logger *zap.Logger) *CachedSpotInstrument {
	return &CachedSpotInstrument{
		baseInstrument: baseInstrument,
		marketsCache:   cache,
		logger:         logger,
	}
}

func (c *CachedSpotInstrument) ViewMarkets(ctx context.Context, rls []roles.UserRole) ([]*domain.Market, error) {
	cachedMarkets, err := c.marketsCache.Get(ctx, rls)
	if err != nil && err != cache.ErrMissed {
		c.logger.Warn("failed to get markets from cache", zap.Error(err))
	}

	if err == nil {
		c.logger.Debug("cache hit for markets", zap.Strings("roles", roles.MapSliceToStrings(rls)))

		return cachedMarkets, nil
	}

	c.logger.Debug("cache miss for markets", zap.Strings("roles", roles.MapSliceToStrings(rls)))

	markets, err := c.baseInstrument.ViewMarkets(ctx, rls)
	if err != nil {
		return nil, err
	}

	err = c.marketsCache.Set(ctx, rls, markets)
	if err != nil {
		c.logger.Warn("failed to cache markets", zap.Error(err))
	}

	return markets, nil
}
