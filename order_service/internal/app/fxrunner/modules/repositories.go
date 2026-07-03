package modules

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/logger"
	shared_retry "github.com/nullableocean/grpcservices/shared/retry"
	"go.uber.org/fx"
)

func PostgresModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, cfg *config.Config) (*pgxpool.Pool, error) {
				pgCnf, err := pgxpool.ParseConfig(cfg.Postgres.DSN)
				if err != nil {
					return nil, fmt.Errorf("failed parse pg dsn: %w", err)
				}
				pgCnf.MaxConns = cfg.Postgres.MaxConns
				pgCnf.MinConns = cfg.Postgres.MinConns
				pgCnf.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
				pgCnf.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime
				pgCnf.ConnConfig.ConnectTimeout = cfg.Postgres.ConnTimeout

				pool, err := pgxpool.NewWithConfig(context.Background(), pgCnf)
				if err != nil {
					return nil, fmt.Errorf("failed create pg pool: %w", err)
				}
				if err := pool.Ping(context.Background()); err != nil {
					pool.Close()
					return nil, fmt.Errorf("failed ping pg: %w", err)
				}
				return pool, nil
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, pool *pgxpool.Pool) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						pool.Close()
						return nil
					},
				})
			},
		),
	)
}

func RepositoriesModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func() *outbox.OutboxWriter {
				return outbox.NewOutboxWriter()
			},
		),
		fx.Provide(
			func(logger *logger.CtxZapLogger, cfg *config.Config, pool *pgxpool.Pool, writer *outbox.OutboxWriter) (*postgres.OrderRepository, error) {
				repoCfg := postgres.RepositoryConfig{
					Retries:          cfg.Repo.MaxRetries,
					RetryBackoffFunc: shared_retry.NewExponentialBackoffFunc(cfg.Repo.BackoffStartDelay, cfg.Repo.BackoffMaxDelay),
				}

				return postgres.NewOrderRepository(logger, pool, writer, repoCfg)
			},
		),
	)
}
