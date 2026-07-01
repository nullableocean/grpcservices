package modules

import (
	"github.com/IBM/sarama"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/bus"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/kafka"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/rdb"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/retry"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type KafkaPublishers struct {
	Updates ports.EventPublisher
	Created ports.EventPublisher
}

func EventPublishersModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func() *bus.EventPublisherBus {
				return bus.NewEventPublisherBus()
			},
		),
		fx.Provide(
			func(logger *zap.Logger, cfg *config.Config, producer sarama.SyncProducer, kafkaMetrics *metrics.KafkaMetricsRecorder) (*KafkaPublishers, error) {
				updates := kafka.NewKafkaPublisher(logger, producer, cfg.Kafka.TopicUpdates, kafkaMetrics)
				created := kafka.NewKafkaPublisher(logger, producer, cfg.Kafka.TopicCreated, kafkaMetrics)
				dlq := kafka.NewKafkaPublisher(logger, producer, cfg.Kafka.DLQTopic, kafkaMetrics)

				dlqCfg := kafka.Config{
					MaxAttempts: cfg.Kafka.ProducerRetries,
					BackoffFunc: retry.NewExponentialBackoffFunc(cfg.Kafka.ProducerStartRetryDelay, cfg.Kafka.ProducerMaxRetryDelay),
				}
				updDlq, err := kafka.NewDlqPublisherDecorator(logger, dlq, updates, kafkaMetrics, dlqCfg)
				if err != nil {
					return nil, err
				}
				crDlq, err := kafka.NewDlqPublisherDecorator(logger, dlq, created, kafkaMetrics, dlqCfg)
				if err != nil {
					return nil, err
				}

				return &KafkaPublishers{
					Updates: updDlq,
					Created: crDlq,
				}, nil
			},
		),
		fx.Provide(
			func(logger *zap.Logger, redisClient *redis.Client, cfg *config.Config) *rdb.Publisher {
				return rdb.NewRedisPublisher(logger, redisClient, cfg.QueueRedis.UpdatesChannel)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, bus *bus.EventPublisherBus, redisPub *rdb.Publisher, kpubs *KafkaPublishers) {
				bus.Register(model.EVENT_ORDER_UPDATED, redisPub)
				bus.Register(model.EVENT_ORDER_UPDATED, kpubs.Updates)
				bus.Register(model.EVENT_ORDER_CREATED, kpubs.Created)
			},
		),
	)
}
