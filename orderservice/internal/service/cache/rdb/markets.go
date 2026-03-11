package rdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/cache"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	versionKey = "markets.ver"
)

type MarketRedisCache struct {
	client  *redis.Client
	ttl     time.Duration
	version int

	logger *zap.Logger
}

func NewMarketCache(l *zap.Logger, rds *redis.Client, ttl time.Duration) *MarketRedisCache {
	return &MarketRedisCache{
		client: rds,
		ttl:    ttl,
		logger: l,
	}
}

func (c *MarketRedisCache) Invalidate(ctx context.Context) error {
	newVer, err := c.client.Incr(ctx, versionKey).Result()
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}

	c.logger.Info("cache version incremented", zap.Int64("new_version", newVer))

	return nil
}

func (c *MarketRedisCache) Get(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	version, err := c.getCurrentVersion(ctx)
	if err != nil {
		c.logger.Error("failed get cache version", zap.Error(err))

		return nil, err
	}

	key := c.createKey(roles, version)

	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrMissed
	}
	if err != nil {
		c.logger.Error("failed get values from redis", zap.Error(err))

		return nil, err
	}

	var markets []*domain.Market
	err = json.Unmarshal(val, &markets)

	if err != nil {
		c.logger.Error("failed json unmarshal", zap.Error(err))

		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return markets, nil
}

func (c *MarketRedisCache) Set(ctx context.Context, rls []roles.UserRole, markets []*domain.Market) error {
	version, err := c.getCurrentVersion(ctx)
	if err != nil {
		c.logger.Error("failed get cache version", zap.Error(err))

		return err
	}

	key := c.createKey(rls, version)

	val, err := json.Marshal(markets)
	if err != nil {
		c.logger.Error("failed json marshal markets", zap.Error(err))

		return fmt.Errorf("json marshal error: %w", err)
	}

	err = c.client.Set(ctx, key, val, c.ttl).Err()
	if err != nil {
		c.logger.Error("failed set values to redis", zap.Error(err))

		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (c *MarketRedisCache) Delete(ctx context.Context, rls []roles.UserRole) error {
	version, err := c.getCurrentVersion(ctx)
	if err != nil {
		c.logger.Error("failed get cache version", zap.Error(err))

		return err
	}

	key := c.createKey(rls, version)
	err = c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("failed delete values from redis", zap.Error(err))
	}

	return nil
}

func (c *MarketRedisCache) getCurrentVersion(ctx context.Context) (int64, error) {
	val, err := c.client.Get(ctx, versionKey).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get version: %w", err)
	}
	return val, nil
}

func (c *MarketRedisCache) createKey(rls []roles.UserRole, version int64) string {
	rlsCopy := make([]roles.UserRole, len(rls))
	copy(rlsCopy, rls)

	roles.SortRolesDesc(rlsCopy)
	rolesString := strings.Join(roles.MapSliceToStrings(rlsCopy), ",")

	return fmt.Sprintf("markets:v%d:%s", version, rolesString)
}
