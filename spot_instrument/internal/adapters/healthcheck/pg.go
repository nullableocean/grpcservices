package healthcheck

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/shared/health"
)

func PostgresHealthcheck(pool *pgxpool.Pool) health.HealthCheck {
	return health.NewHealthCheck("postgres", func(ctx context.Context) error {
		return pool.Ping(ctx)
	})
}
