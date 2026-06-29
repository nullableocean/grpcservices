package ports

import (
	"context"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type OrderRateLimiter interface {
	Check(ctx context.Context, user model.User) error
	Rollback(ctx context.Context, user model.User) error
}

type CacheRateLimitCounter interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
	Decrement(ctx context.Context, key string) error
}
