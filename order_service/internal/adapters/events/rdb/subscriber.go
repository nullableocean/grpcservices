package rdb

import (
	"context"
	"errors"
	"sync"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrStopped = errors.New("service stopped")
)

type RedisEventSubscriber struct {
	handler  ports.MessageHandler
	client   *redis.Client
	pubsub   *redis.PubSub
	channels []string

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	stopped bool
	started bool

	logger *logger.CtxZapLogger
}

func NewRedisSubscriber(logger *logger.CtxZapLogger, client *redis.Client, channels []string, handler ports.MessageHandler) *RedisEventSubscriber {
	return &RedisEventSubscriber{
		handler:  handler,
		client:   client,
		channels: channels,
		logger:   logger,
	}
}

func (s *RedisEventSubscriber) Start(ctx context.Context) error {
	if s.stopped {
		return ErrStopped
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.pubsub = s.client.Subscribe(ctx, s.channels...)
	if err := s.pubsub.Ping(ctx); err != nil {
		return errors.Join(err, s.pubsub.Close())
	}

	s.wg.Add(1)
	go s.startReceive(ctx)

	s.started = true
	return nil
}

func (s *RedisEventSubscriber) startReceive(ctx context.Context) {
	defer s.wg.Done()

	msgCh := s.pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgCh:
			err := s.handler.Handle(ctx, msg)
			if err != nil {
				s.logger.Error(ctx, "failed handling event from redis queue", zap.String("channel", msg.Channel), zap.Error(err))
			}
		}
	}
}

func (s *RedisEventSubscriber) Stop() error {
	if !s.started {
		return nil
	}

	s.cancel()
	s.wg.Wait()
	s.stopped = true

	return s.pubsub.Close()
}
