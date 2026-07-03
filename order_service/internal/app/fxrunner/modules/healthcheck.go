package modules

import (
	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/healthcheck"
	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type ModulesReadyHealths struct {
	fx.In

	OutboxHealth *OutboxHealth
	Grpc         *GrpcServerHealth
	RedisSub     *RedisSubHealth
}

func HealthcheckModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, kafkaClient sarama.Client, pgpool *pgxpool.Pool, rdb *redis.Client, healths ModulesReadyHealths) *health.HealthCheker {
				return health.NewHealthCheker(
					healthcheck.KafkaHealthcheck(logger, kafkaClient),
					healthcheck.PostgresHealthcheck(pgpool),
					healthcheck.RedisChecker(rdb),
					healths.OutboxHealth.HealthCheck(),
					healths.Grpc.HealthCheck(),
					healths.RedisSub.HealthCheck(),
				)
			},
		),
	)
}
