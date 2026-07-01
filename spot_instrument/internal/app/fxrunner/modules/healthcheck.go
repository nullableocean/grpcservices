package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/healthcheck"
	"go.uber.org/fx"
)

func HealthModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(pgpool *pgxpool.Pool, grpcHealth *GrpcServerHealth) *health.HealthCheker {
				return health.NewHealthCheker(
					healthcheck.PostgresHealthcheck(pgpool),
					grpcHealth.HealthCheck(),
				)
			},
		),
	)
}
