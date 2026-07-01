package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/retry"
	"go.uber.org/fx"
)

func KafkaProducerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg *config.Config) (sarama.SyncProducer, sarama.Client, error) {
				saramaCfg := sarama.NewConfig()
				ver, err := sarama.ParseKafkaVersion(cfg.Kafka.Version)
				if err != nil {
					return nil, nil, err
				}
				saramaCfg.Version = ver

				switch cfg.Kafka.ProducerAcks {
				case "all", "-1":
					saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
				case "one":
					saramaCfg.Producer.RequiredAcks = sarama.WaitForLocal
				default:
					saramaCfg.Producer.RequiredAcks = sarama.NoResponse
				}

				switch cfg.Kafka.ProducerCompression {
				case "snappy":
					saramaCfg.Producer.Compression = sarama.CompressionSnappy
				case "gzip":
					saramaCfg.Producer.Compression = sarama.CompressionGZIP
				case "lz4":
					saramaCfg.Producer.Compression = sarama.CompressionLZ4
				default:
					saramaCfg.Producer.Compression = sarama.CompressionNone
				}

				if cfg.Kafka.ProducerRetries > 0 {
					backoffFunc := retry.NewExponentialBackoffFunc(cfg.Kafka.ProducerStartRetryDelay, cfg.Kafka.ProducerMaxRetryDelay)
					saramaCfg.Producer.Retry.Max = cfg.Kafka.ProducerRetries
					saramaCfg.Producer.Retry.BackoffFunc = func(attempts, maxAttempts int) time.Duration {
						return backoffFunc(attempts)
					}
				}

				saramaCfg.Producer.Flush.Messages = cfg.Kafka.ProducerBatchMessages
				saramaCfg.Producer.Flush.Bytes = cfg.Kafka.ProducerMaxMessageBytes
				saramaCfg.Producer.Flush.Frequency = cfg.Kafka.ProducerBatchFlushFrequency
				saramaCfg.Net.DialTimeout = cfg.Kafka.DialTimeout
				saramaCfg.Net.WriteTimeout = cfg.Kafka.WriteTimeout
				if cfg.Kafka.AutoTopicCreation {
					saramaCfg.Metadata.AllowAutoTopicCreation = true
				}
				saramaCfg.Producer.Return.Successes = true
				saramaCfg.Producer.Return.Errors = true

				client, err := sarama.NewClient(cfg.Kafka.Brokers, saramaCfg)
				if err != nil {
					return nil, nil, err
				}
				if err := client.RefreshMetadata(); err != nil {
					client.Close()
					return nil, nil, fmt.Errorf("refresh metadata failed: %w", err)
				}

				producer, err := sarama.NewSyncProducerFromClient(client)
				return producer, client, err
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, producer sarama.SyncProducer, client sarama.Client) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						err := producer.Close()
						if err != nil {
							return fmt.Errorf("failed close kafka producer: %w", err)
						}

						err = client.Close()
						if err != nil {
							return fmt.Errorf("failed close kafka client: %w", err)
						}

						return nil
					},
				})
			},
		),
	)
}
