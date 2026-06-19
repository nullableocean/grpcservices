package app

import (
	"fmt"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/access"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/cache/rdb"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/bus"
	kafka_publisher "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/kafka"
	events_rdb "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/rdb"
	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/client"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	shared_retry "github.com/nullableocean/grpcservices/shared/retry"
)

func (a *App) initServices() error {
	spotProtoClient := spotv1.NewSpotInstrumentClient(a.spotConn)
	spotInstrument, err := client.NewSpotInstrumentClient(a.logger, spotProtoClient, client.Option{
		RequestTimeout: a.cnf.GRPC.ClientTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed create spot instrument client: %w", err)
	}

	outboxWriter := outbox.NewOutboxWriter()
	repoCfg := postgres.RepositoryConfig{
		Retries:          int(a.cnf.Repo.MaxRetries),
		RetryBackoffFunc: shared_retry.NewExponentialBackoffFunc(a.cnf.Repo.BackoffStartDelay, a.cnf.Repo.BackoffMaxDelay),
	}
	orderRepo, err := postgres.NewOrderRepository(a.logger, a.pgPool, outboxWriter, repoCfg)
	if err != nil {
		return fmt.Errorf("failed create order repository: %w", err)
	}

	accessService := access.NewRoleAccessService()
	metricsRecorder := metrics.NewOrderMetricsRecorder(a.metricsReg)
	redisRecorder := metrics.NewRedisMetricsRecorder(a.metricsReg)
	a.orderService = order.NewOrderService(
		a.logger,
		orderRepo,
		spotInstrument,
		accessService,
		metricsRecorder,
		rdb.NewRedisIdempotencyCache(a.redis, a.cnf.Cache.TTL, redisRecorder),
	)

	pubBus := bus.NewEventPublisherBus()

	// GRPC STREAM NOTIFIER

	a.updatesNotifier, err = updatenotifier.NewUpdateNotifier(a.logger, updatenotifier.Options{
		SendTimeoutOnSub: a.cnf.Events.STREAM_SEND_TIMEOUT,
		SendTries:        a.cnf.Events.STREAM_SEND_RETRIES,
	})
	if err != nil {
		return fmt.Errorf("failed create update notifier: %w", err)
	}

	// REDIS EVENTS

	updatesHandler := events_rdb.NewUpdatesMessageHandler(a.logger, a.updatesNotifier)
	a.redisSub = events_rdb.NewRedisSubscriber(a.logger, a.redis, []string{a.cnf.QueueRedis.UpdatesChannel}, updatesHandler)
	a.closers = append(a.closers, func() error {
		if err := a.redisSub.Stop(); err != nil {
			return fmt.Errorf("failed stop redis subsriber: %w", err)
		}

		return nil
	})

	redisPublisher := events_rdb.NewRedisPublisher(a.logger, a.redis, a.cnf.QueueRedis.UpdatesChannel)

	// KAFKA PUBLISHERS

	kafkaMetrics := metrics.NewKafkaMetricsRecorder(a.metricsReg)
	kafkaProducer, err := newSaramaSyncProducer(a.cnf.Kafka)
	if err != nil {
		return fmt.Errorf("failed to create Sarama producer: %w", err)
	}
	a.closers = append(a.closers, func() error {
		if err := kafkaProducer.Close(); err != nil {
			return fmt.Errorf("failed close kafka producer: %w", err)
		}
		return nil
	})

	kafkaUpdatedPublisher := kafka_publisher.NewKafkaPublisher(a.logger, kafkaProducer, a.cnf.Kafka.TopicUpdates, kafkaMetrics)
	kafkaCreatedPublisher := kafka_publisher.NewKafkaPublisher(a.logger, kafkaProducer, a.cnf.Kafka.TopicCreated, kafkaMetrics)
	dlqPublisher := kafka_publisher.NewKafkaPublisher(a.logger, kafkaProducer, a.cnf.Kafka.DLQTopic, kafkaMetrics)

	dlqPublishCfg := kafka_publisher.Config{
		MaxAttempts: a.cnf.Kafka.ProducerRetries,
		BackoffFunc: shared_retry.NewExponentialBackoffFunc(a.cnf.Kafka.ProducerStartRetryDelay, a.cnf.Kafka.ProducerMaxRetryDelay),
	}

	publisherUpdatesDlqDecorator, err := kafka_publisher.NewDlqPublisherDecorator(a.logger, dlqPublisher, kafkaUpdatedPublisher, kafkaMetrics, dlqPublishCfg)
	if err != nil {
		return fmt.Errorf("failed create DlqPublishRetrayer for updates events: %w", err)
	}
	publisherCreatedDlqDecorator, err := kafka_publisher.NewDlqPublisherDecorator(a.logger, dlqPublisher, kafkaCreatedPublisher, kafkaMetrics, dlqPublishCfg)
	if err != nil {
		return fmt.Errorf("failed create DlqPublishRetrayer for created events: %w", err)
	}

	// REGISTER PUBLISHERS

	pubBus.Register(model.EVENT_ORDER_UPDATED, redisPublisher)
	pubBus.Register(model.EVENT_ORDER_UPDATED, publisherUpdatesDlqDecorator)
	pubBus.Register(model.EVENT_ORDER_CREATED, publisherCreatedDlqDecorator)

	outboxMetrics := metrics.NewOutboxMetricsRecorder(a.metricsReg)
	outboxRelay, err := outbox.NewRelay(a.logger, a.pgPool, pubBus, outboxMetrics, outbox.Config{
		Interval:     a.cnf.Outbox.PollInterval,
		BatchSize:    a.cnf.Outbox.BatchSize,
		BatchTimeout: a.cnf.Outbox.BatchHandleTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed create outbox relay: %w", err)
	}
	a.closers = append(a.closers, func() error {
		err := outboxRelay.Stop()
		if err != nil {
			return fmt.Errorf("failed stop outbox relay: %w", err)
		}
		return nil
	})

	a.outboxRelay = outboxRelay

	return nil
}
