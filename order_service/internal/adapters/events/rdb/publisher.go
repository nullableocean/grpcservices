package rdb

import (
	"context"
	"encoding/json"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var _ ports.EventPublisher = &Publisher{}

type Publisher struct {
	client  *redis.Client
	channel string

	logger *logger.CtxZapLogger
}

func NewRedisPublisher(logger *logger.CtxZapLogger, client *redis.Client, channel string) *Publisher {
	return &Publisher{
		client:  client,
		channel: channel,
		logger:  logger,
	}
}

func (p *Publisher) Publish(ctx context.Context, event model.Event) error {
	logger := p.logger.With(zap.String("event_uuid", event.ID()))

	bdata, err := json.Marshal(event)
	if err != nil {
		logger.Error(ctx, "failed to json marshal event", zap.Error(err))
		return err
	}

	err = p.client.Publish(ctx, p.channel, bdata).Err()
	if err != nil {
		logger.Error(ctx, "failed to publish event", zap.Error(err))
		return err
	}

	logger.Debug(ctx, "event published to redis")

	return nil
}
