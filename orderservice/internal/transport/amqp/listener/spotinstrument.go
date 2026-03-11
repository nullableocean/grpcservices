package listener

import (
	"context"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MarketCache interface {
	Invalidate(ctx context.Context) error
}

type SpotInstrumentUpdateListener struct {
	kread *kafka.Reader
	cache MarketCache

	logger *zap.Logger
}

func NewSpotInstrumentUpdateListener(logger *zap.Logger, kr *kafka.Reader, cache MarketCache) *SpotInstrumentUpdateListener {
	return &SpotInstrumentUpdateListener{
		kread:  kr,
		cache:  cache,
		logger: logger,
	}
}

func (l *SpotInstrumentUpdateListener) StartListen(ctx context.Context) error {
	l.logger.Info("starting spot instrument update listener", zap.String("topic", l.kread.Config().Topic))

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("listener stopped by context")
			return ctx.Err()
		default:
		}

		msg, err := l.kread.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				l.logger.Info("listener context done", zap.Error(err))
				return err
			}
			l.logger.Error("failed to fetch message from Kafka", zap.Error(err))
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := l.cache.Invalidate(ctx); err != nil {
			l.logger.Error("failed to invalidate cache", zap.Error(err))
			continue
		}

		if err := l.kread.CommitMessages(ctx, msg); err != nil {
			l.logger.Error("failed to commit offset", zap.Error(err))
			continue
		}

		l.logger.Info("cache invalidated after market update",
			zap.String("topic", msg.Topic),
			zap.Int64("offset", msg.Offset),
		)
	}
}
