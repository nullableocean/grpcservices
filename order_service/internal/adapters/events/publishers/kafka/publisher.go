package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Publisher struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewKafkaPublisher(logger *zap.Logger, writer *kafka.Writer) *Publisher {
	return &Publisher{
		writer: writer,
		logger: logger,
	}
}

func (p *Publisher) Publish(ctx context.Context, event model.Event) error {
	payload, err := event.Payload()
	if err != nil {
		p.logger.Error("failed to serialize event data",
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		return fmt.Errorf("failed event data: %w", err)
	}

	key := []byte(event.OrderID())
	msg := kafka.Message{
		Key:   key,
		Value: payload,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType().String())},
			{Key: "event_id", Value: []byte(event.ID())},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("failed to write message to kafka",
			zap.String("topic", p.writer.Topic),
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		return fmt.Errorf("failed write to kafka: %w", err)
	}

	p.logger.Info("event published in kafka",
		zap.String("topic", p.writer.Topic),
		zap.String("event_id", event.ID()),
		zap.String("order_id", event.OrderID()),
		zap.String("event_type", event.EventType().String()),
	)

	return nil
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
