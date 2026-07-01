package modules

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func RedisModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg *config.Config) (*redis.Client, error) {
				client := redis.NewClient(&redis.Options{
					Addr:            cfg.Cache.RedisAddr,
					Password:        cfg.Cache.RedisPassword,
					DB:              cfg.Cache.RedisDB,
					DialTimeout:     cfg.Cache.DialTimeout,
					ReadTimeout:     cfg.Cache.ReadTimeout,
					WriteTimeout:    cfg.Cache.WriteTimeout,
					MaxRetries:      cfg.Cache.MaxRetries,
					MinRetryBackoff: cfg.Cache.MinBackoffDelay,
					MaxRetryBackoff: cfg.Cache.MaxBackoffDelay,
				})
				if err := client.Ping(context.Background()).Err(); err != nil {
					return nil, fmt.Errorf("failed redis ping: %w", err)
				}
				return client, nil
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, client *redis.Client) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return client.Close()
					},
				})
			},
		),
	)
}
