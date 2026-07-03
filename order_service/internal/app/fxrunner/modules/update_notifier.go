package modules

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/rdb"
	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func UpdateNotifierModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, cfg *config.Config) (*updatenotifier.UpdateNotifier, error) {
				return updatenotifier.NewUpdateNotifier(logger, updatenotifier.Options{
					SendTimeoutOnSub: cfg.Events.StreamSendTimeout,
					SendTries:        cfg.Events.StreamSendRetries,
				})
			},
		),
	)
}

type RedisSubHealth struct {
	ready atomic.Bool
}

func (o *RedisSubHealth) HealthCheck() health.HealthCheck {
	return health.NewHealthCheck("redis_subscriber", func(ctx context.Context) error {
		if !o.ready.Load() {
			return errors.New("redis_subscriber not ready")
		}
		return nil
	})
}

func RedisSubsriberModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, cfg *config.Config, redisClient *redis.Client, notifier *updatenotifier.UpdateNotifier) (*rdb.RedisEventSubscriber, *RedisSubHealth) {
				handler := rdb.NewUpdatesMessageHandler(logger, notifier)
				return rdb.NewRedisSubscriber(logger, redisClient, []string{cfg.QueueRedis.UpdatesChannel}, handler), &RedisSubHealth{}
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, logger *zap.Logger, sub *rdb.RedisEventSubscriber, health *RedisSubHealth) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						health.ready.Store(true)
						go func() {
							if err := sub.Start(ctx); err != nil {
								logger.Error("redis subscriber error", zap.Error(err))
								health.ready.Store(false)
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("stopping redis subscriber...")
						health.ready.Store(true)
						return sub.Stop()
					},
				})
			},
		),
	)
}
