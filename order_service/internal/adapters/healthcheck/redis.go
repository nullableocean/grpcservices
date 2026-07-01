package healthcheck

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/redis/go-redis/v9"
)

func RedisChecker(client *redis.Client) health.HealthCheck {
	return health.NewHealthCheck("redis", func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	})
}
