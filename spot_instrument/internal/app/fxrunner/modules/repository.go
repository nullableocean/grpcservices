package modules

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func PostgresModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg *config.Config) (*pgxpool.Pool, error) {
				pgCnf, err := pgxpool.ParseConfig(cfg.Postgres.DSN)
				if err != nil {
					return nil, fmt.Errorf("parse pg dsn: %w", err)
				}

				pgCnf.MaxConns = cfg.Postgres.MaxConns
				pgCnf.MinConns = cfg.Postgres.MinConns
				pgCnf.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
				pgCnf.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime
				pgCnf.ConnConfig.ConnectTimeout = cfg.Postgres.ConnTimeout

				pool, err := pgxpool.NewWithConfig(context.Background(), pgCnf)
				if err != nil {
					return nil, fmt.Errorf("create pg pool: %w", err)
				}

				if err := pool.Ping(context.Background()); err != nil {
					pool.Close()
					return nil, fmt.Errorf("ping pg: %w", err)
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
			func(ctxLogger *logger.CtxZapLogger, pool *pgxpool.Pool) (*postgres.MarketRepository, error) {
				return postgres.NewMarketRepository(ctxLogger, pool)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, repo *postgres.MarketRepository, cfg *config.Config, logger *zap.Logger) {
				var cancel context.CancelFunc
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						ctx, cancel = context.WithCancel(ctx)
						repo.StartRefreshingRoles(ctx, cfg.SpotRepo.RolesRefreshInterval)
						logger.Info("started market roles refresher")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						if cancel != nil {
							cancel()
						}
						repo.Stop()
						logger.Info("stopped market roles refresher")
						return nil
					},
				})
			},
		),
	)
}
