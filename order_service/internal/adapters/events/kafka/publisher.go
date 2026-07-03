package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"go.uber.org/zap"
)

var _ ports.EventPublisher = &Publisher{}

type Publisher struct {
	producer sarama.SyncProducer
	topic    string
	metrics  *metrics.KafkaMetricsRecorder
	logger   *logger.CtxZapLogger
}

func NewKafkaPublisher(logger *logger.CtxZapLogger, producer sarama.SyncProducer, topic string, metrics *metrics.KafkaMetricsRecorder) *Publisher {
	return &Publisher{
		producer: producer,
		topic:    topic,
		logger:   logger,
		metrics:  metrics,
	}
}

func (p *Publisher) Publish(ctx context.Context, event model.Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("grpc request panic: %v", r)
			p.logger.Error(ctx, "panic in kafka publisher", zap.Any("error", msg))

			err = fmt.Errorf("panic in publisher: %w: %s", errs.ErrInternal, msg)
		}
	}()

	payload, err := event.Payload()
	if err != nil {
		p.logger.Error(ctx, "failed to serialize event data",
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)
		return fmt.Errorf("failed event data: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.StringEncoder(event.GetOrderUUID()),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
		Headers:   p.getHeaders(event),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.metrics.MessagePublishFailed(ctx)
		p.logger.Error(ctx, "failed to write message to kafka",
			zap.String("topic", p.topic),
			zap.String("event_id", event.ID()),
			zap.Error(err),
		)

		return fmt.Errorf("failed write to kafka: %w", err)
	}

	p.metrics.MessagePublished(ctx)

	p.logger.Debug(ctx, "event published in kafka",
		zap.String("topic", p.topic),
		zap.String("event_id", event.ID()),
		zap.String("order_id", event.GetOrderUUID()),
		zap.String("event_type", event.EventType().String()),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

func (p *Publisher) getHeaders(event model.Event) []sarama.RecordHeader {
	xreq := xrequestid.NewXRequestId()

	return []sarama.RecordHeader{
		{Key: []byte("event_type"), Value: []byte(event.EventType().String())},
		{Key: []byte("event_id"), Value: []byte(event.ID())},
		{Key: []byte(xrequestid.XREQUEST_ID_KEY), Value: []byte(xreq)},
	}
}
