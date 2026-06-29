package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var _ ports.CacheRateLimitCounter = &RateLimitCache{}

type RateLimitCache struct {
	client  *redis.Client
	metrics *metrics.RedisMetricsRecorder
	logger  *zap.Logger
}

func NewRedisRateLimitCache(logger *zap.Logger, client *redis.Client, metrics *metrics.RedisMetricsRecorder) *RateLimitCache {
	return &RateLimitCache{
		client:  client,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *RateLimitCache) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		c.metrics.CacheSetError(ctx)

		return 0, fmt.Errorf("redis incr failed: %w", err)
	}

	if val == 1 {
		if err := c.client.Expire(ctx, key, window).Err(); err != nil {
			c.logger.Error("failed to set TTL for rate limit key", zap.String("key", key), zap.Error(err))
			c.metrics.CacheSetError(ctx)
		}
	}

	return val, nil
}

func (c *RateLimitCache) Decrement(ctx context.Context, key string) error {
	_, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		c.metrics.CacheSetError(ctx)
		return fmt.Errorf("redis decr failed: %w", err)
	}

	return nil
}
