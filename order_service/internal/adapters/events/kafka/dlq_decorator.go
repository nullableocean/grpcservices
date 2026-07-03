package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.uber.org/zap"
)

var _ ports.EventPublisher = &DlqPublisherDecorator{}

type RetryBackoffFunc func(attempt int) time.Duration

type DlqPublisherDecorator struct {
	dlqPublisher *Publisher

	publisher *Publisher
	logger    *logger.CtxZapLogger
	metrics   *metrics.KafkaMetricsRecorder

	maxAttemps  int
	backoffFunc RetryBackoffFunc
}

type Config struct {
	MaxAttempts int
	BackoffFunc RetryBackoffFunc
}

func NewDlqPublisherDecorator(logger *logger.CtxZapLogger, dlqWriter *Publisher, publisher *Publisher, metrics *metrics.KafkaMetricsRecorder, opts Config) (*DlqPublisherDecorator, error) {
	if opts.MaxAttempts <= 0 {
		return nil, fmt.Errorf("invalid attempts for DlqPublishRetrayer: %d", opts.MaxAttempts)
	}

	return &DlqPublisherDecorator{
		dlqPublisher: dlqWriter,
		publisher:    publisher,
		logger:       logger,
		metrics:      metrics,
		maxAttemps:   opts.MaxAttempts,
		backoffFunc:  opts.BackoffFunc,
	}, nil
}

func (p *DlqPublisherDecorator) Publish(ctx context.Context, event model.Event) error {
	var lastErr error
	for attempt := 0; attempt < p.maxAttemps; attempt++ {
		err := p.publisher.Publish(ctx, event)
		if err == nil {
			return nil
		}

		p.logger.Warn(ctx, "failed to publish event, will retry",
			zap.Int("attempt", attempt+1),
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		lastErr = err
		backoff := p.backoffFunc(attempt)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			p.logger.Error(ctx, "context cancelled during retry",
				zap.String("event_id", event.ID()),
			)

			return ctx.Err()
		}
	}

	p.logger.Error(ctx, "all retires done. sending to DLQ",
		zap.String("event_id", event.ID()),
		zap.Error(lastErr),
	)

	if err := p.dlqPublisher.Publish(ctx, event); err != nil {
		p.logger.Error(ctx, "failed to publish to DLQ",
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		p.metrics.MessagePublishFailed(ctx)
		return fmt.Errorf("main publish failed. DLQ publish also failed: %w", err)
	}

	p.metrics.MessageSendToDlq(ctx)
	return fmt.Errorf("event sent to DLQ after %d attempts: %w", p.maxAttemps, lastErr)
}
