package modules

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/bus"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type OutboxHealth struct {
	ready atomic.Bool
}

func (o *OutboxHealth) HealthCheck() health.HealthCheck {
	return health.NewHealthCheck("outbox", func(ctx context.Context) error {
		if !o.ready.Load() {
			return errors.New("outbox not ready")
		}
		return nil
	})
}

func OutboxModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, cfg *config.Config, pool *pgxpool.Pool, bus *bus.EventPublisherBus, outboxMetrics *metrics.OutboxMetricsRecorder) (*outbox.OutboxRelay, *OutboxHealth, error) {
				outbox, err := outbox.NewRelay(logger, pool, bus, outboxMetrics, outbox.Config{
					Interval:     cfg.Outbox.PollInterval,
					BatchSize:    cfg.Outbox.BatchSize,
					BatchTimeout: cfg.Outbox.BatchHandleTimeout,
					MaxRetry:     cfg.Outbox.MaxRetries,
				})

				return outbox, &OutboxHealth{}, err
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, logger *zap.Logger, relay *outbox.OutboxRelay, health *OutboxHealth) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go relay.Start(ctx)
						health.ready.Store(true)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("stopping outbox relay...")
						health.ready.Store(false)

						return relay.Stop()
					},
				})
			},
		),
	)
}
