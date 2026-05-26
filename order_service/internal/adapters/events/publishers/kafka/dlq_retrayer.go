package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.uber.org/zap"
)

type RetryBackoffFunc func(attempt int) time.Duration

type DlqPublishRetrayer struct {
	dlqPublisher *Publisher

	publisher *Publisher
	logger    *zap.Logger
	metrics   *metrics.KafkaMetricsRecorder

	maxAttemps  int
	backoffFunc RetryBackoffFunc
}

type Config struct {
	MaxAttempts int
	BackoffFunc RetryBackoffFunc
}

func NewDlqPublishRetrayer(logger *zap.Logger, dlqWriter *Publisher, publisher *Publisher, metrics *metrics.KafkaMetricsRecorder, opts Config) (*DlqPublishRetrayer, error) {
	if opts.MaxAttempts <= 0 {
		return nil, fmt.Errorf("invalid attempts for DlqPublishRetrayer: %d", opts.MaxAttempts)
	}

	return &DlqPublishRetrayer{
		dlqPublisher: dlqWriter,
		publisher:    publisher,
		logger:       logger,
		metrics:      metrics,
		maxAttemps:   opts.MaxAttempts,
		backoffFunc:  opts.BackoffFunc,
	}, nil
}

func (p *DlqPublishRetrayer) Publish(ctx context.Context, event model.Event) error {
	var lastErr error
	for attempt := 0; attempt < p.maxAttemps; attempt++ {
		err := p.publisher.Publish(ctx, event)
		if err == nil {
			return nil
		}

		p.logger.Warn("failed to publish event, will retry",
			zap.Int("attempt", attempt+1),
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		lastErr = err
		backoff := p.backoffFunc(attempt)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			p.logger.Error("context cancelled during retry",
				zap.String("event_id", event.ID()),
			)

			return ctx.Err()
		}
	}

	p.logger.Error("all retires done. sending to DLQ",
		zap.String("event_id", event.ID()),
		zap.Error(lastErr),
	)

	if err := p.dlqPublisher.Publish(ctx, event); err != nil {
		p.logger.Error("failed to publish to DLQ",
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		p.metrics.MessagePublishFailed(ctx)
		return fmt.Errorf("main publish failed. DLQ publish also failed: %w", err)
	}

	p.metrics.MessageSendToDlq(ctx)
	return fmt.Errorf("event sent to DLQ after %d attempts: %w", p.maxAttemps, lastErr)
}
