package healthcheck

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/nullableocean/grpcservices/shared/health"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.uber.org/zap"
)

// KafkaHealthcheck проверяет коннекты к брокерам кафки
func KafkaHealthcheck(logger *logger.CtxZapLogger, client sarama.Client) health.HealthCheck {
	return health.NewHealthCheck("kafka", func(ctx context.Context) error {
		if err := client.RefreshMetadata(); err != nil {
			logger.Error(ctx, "failed to refresh kafka metadata", zap.Error(err))
			return fmt.Errorf("refresh metadata failed: %w", err)
		}

		brokers := client.Brokers()
		if len(brokers) == 0 {
			logger.Error(ctx, "kafka empty brokers")
			return errors.New("empty kafka brokers")
		}

		connected := 0
		for _, broker := range brokers {
			ok, err := broker.Connected()
			if err != nil {
				logger.Error(ctx, "failed kafka check connect", zap.Error(err), zap.String("broker_addr", broker.Addr()))
			}
			if ok {
				connected++
			}
		}

		if connected == 0 {
			return errors.New("no one kafka brokers dont connected")
		}

		return nil
	})
}
