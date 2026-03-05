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
)

type MarketRedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewMarketCache(rds *redis.Client, ttl time.Duration) *MarketRedisCache {
	return &MarketRedisCache{
		client: rds,
		ttl:    ttl,
	}
}

func (c *MarketRedisCache) Get(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	key := c.getKey(roles)

	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrMissed
	}

	if err != nil {
		return nil, err
	}

	var markets []*domain.Market
	err = json.Unmarshal(val, &markets)

	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return markets, nil
}

func (c *MarketRedisCache) Set(ctx context.Context, roles []roles.UserRole, markets []*domain.Market) error {
	key := c.getKey(roles)

	val, err := json.Marshal(markets)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	err = c.client.Set(ctx, key, val, c.ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (c *MarketRedisCache) Delete(ctx context.Context, roles []roles.UserRole) error {
	key := c.getKey(roles)
	return c.client.Del(ctx, key).Err()
}

func (c *MarketRedisCache) getKey(rls []roles.UserRole) string {
	rlsCopy := make([]roles.UserRole, len(rls))
	copy(rlsCopy, rls)

	roles.SortRolesDesc(rlsCopy)
	roleStrings := roles.MapSliceToStrings(rlsCopy)

	key := "markets:" + strings.Join(roleStrings, ",")
	return key
}
