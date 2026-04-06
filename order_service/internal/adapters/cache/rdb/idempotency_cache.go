package rdb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

var _ ports.IdempotencyCache = &IdempotencyCache{}

const (
	idempotencyPrefix = "idempotency"
)

type IdempotencyCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisIdempotencyCache(client *redis.Client, ttl time.Duration) *IdempotencyCache {
	return &IdempotencyCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *IdempotencyCache) Get(ctx context.Context, key string) (*model.IdempotencyData, error) {
	val, err := r.client.Get(ctx, idempotencyPrefix+key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data model.IdempotencyData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *IdempotencyCache) SetIfNotExist(ctx context.Context, key string, data *model.IdempotencyData) (bool, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	// deprecated
	// ok, err := r.client.SetNX(ctx, idempotencyPrefix+key, b, r.ttl).Result()
	// if err != nil {
	// 	return false, err
	// }

	status, err := r.client.SetArgs(ctx, idempotencyPrefix+key, b, redis.SetArgs{
		Mode: "NX",
		TTL:  r.ttl,
	}).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	return status == "OK", nil
}

func (r *IdempotencyCache) Update(ctx context.Context, key string, data *model.IdempotencyData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, idempotencyPrefix+key, b, r.ttl).Err()
}

func (r *IdempotencyCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, idempotencyPrefix+key).Err()
}
