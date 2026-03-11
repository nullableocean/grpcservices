package listener

import (
	"context"
	"errors"
	"sync"
	"time"

	ordereventsv1 "github.com/nullableocean/grpcservices/api/gen/events/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/outside"
	"github.com/nullableocean/grpcservices/shared/limiter"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	defaultProcLimit = 4
	defaultRetries   = 4
)

type UpdateEventHandler interface {
	Handle(ctx context.Context, update *outside.UpdateStatusEvent) error
}

type UpdateListener struct {
	kafkaReader *kafka.Reader
	dlqWriter   *kafka.Writer

	handler UpdateEventHandler
	logger  *zap.Logger

	maxRetry      int
	handleRetries map[string]int
	mu            sync.Mutex

	processLimiter *limiter.Limiter
}

type Option struct {
	ProcessLimit int
	MaxRetries   int
}

func NewUpdateListener(l *zap.Logger, kread *kafka.Reader, dlqw *kafka.Writer, h UpdateEventHandler, opt Option) *UpdateListener {
	limit := opt.ProcessLimit
	retries := opt.MaxRetries

	if limit <= 0 {
		limit = defaultProcLimit
	}

	if retries < 0 {
		retries = defaultRetries
	}

	return &UpdateListener{
		kafkaReader: kread,
		dlqWriter:   dlqw,
		handler:     h,
		logger:      l,

		maxRetry:      retries,
		handleRetries: make(map[string]int),
		mu:            sync.Mutex{},

		processLimiter: limiter.New(limit),
	}
}

func (l *UpdateListener) StartListen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("stop update listener by context")
			return ctx.Err()
		default:
		}

		l.logger.Info("fetching messages from kafka", zap.String("topic", l.kafkaReader.Config().Topic))
		msg, err := l.kafkaReader.FetchMessage(ctx)
		l.logger.Info("fetched message...")

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				l.logger.Error("close broker listener by context", zap.Error(err))

				return err
			}

			l.logger.Error("failed to fetch message from broker", zap.Error(err))

			time.Sleep(100 * time.Millisecond)
			continue
		}

		err = l.processLimiter.AcquireContext(ctx)
		if err != nil {
			l.logger.Info("stop update listener by context")
			return ctx.Err()
		}

		go l.handleMsg(ctx, msg)
	}
}

func (l *UpdateListener) handleMsg(ctx context.Context, msg kafka.Message) {
	defer l.processLimiter.Release()

	traceCtx, span := l.startTracing(ctx, msg)
	defer span.End()

	msgKey := string(msg.Key)
	reqId := l.getRequestIdFromHeaders(msg.Headers)
	logger := l.logger.With(
		zap.String(xrequestid.XREQUEST_ID_KEY, reqId),
		zap.String("event_key", msgKey),
	)

	logger.Info("got order update event")

	span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqId))

	event, err := l.unmarshalDataToUpdateEvent(msg.Value)
	if err != nil {
		logger.Error("failed unmarshal data", zap.Error(err))
		span.AddEvent("unmarshal error")

		l.kafkaReader.CommitMessages(ctx, msg)
		l.writeToDLQ(ctx, msg, "unmarshal error", err.Error())

		return
	}
	logger = logger.With(zap.String("event_uuid", event.UUID))

	err = l.handler.Handle(traceCtx, event)
	if err != nil && !errors.Is(err, outside.ErrEventAlreadyHandled) {
		logger.Error("failed handle event", zap.Error(err))

		l.mu.Lock()
		tries := l.handleRetries[event.UUID]
		if tries >= l.maxRetry {
			logger.Info("write event to DLQ")
			span.AddEvent("write to DLQ")

			l.kafkaReader.CommitMessages(ctx, msg)
			l.writeToDLQ(ctx, msg, "retries limited", err.Error())

			delete(l.handleRetries, event.UUID)
			l.mu.Unlock()

			return
		}

		l.handleRetries[event.UUID] += 1
		l.mu.Unlock()

		logger.Info("event go to retry")
		span.AddEvent("event retry")
		return
	}

	commitErr := l.kafkaReader.CommitMessages(ctx, msg)
	if commitErr != nil {
		logger.Error("failed to commit after unmarshal error", zap.Error(commitErr))
		span.AddEvent("commit error")
		return
	}

	l.mu.Lock()
	delete(l.handleRetries, event.UUID)
	l.mu.Unlock()

	logger.Info("success event handle")
	span.AddEvent("commit success")
}

func (l *UpdateListener) startTracing(ctx context.Context, msg kafka.Message) (context.Context, trace.Span) {
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.HeaderCarrier{}
	for _, h := range msg.Headers {
		carrier.Set(h.Key, string(h.Value))
	}

	ctx = propagator.Extract(ctx, carrier)
	traceCtx, span := otel.Tracer("order_updates_listener").Start(ctx, "got_update_event")

	return traceCtx, span
}

func (l *UpdateListener) unmarshalDataToUpdateEvent(value []byte) (*outside.UpdateStatusEvent, error) {
	protoUpdateEvent := ordereventsv1.UpdateStatus{}

	err := proto.Unmarshal(value, &protoUpdateEvent)
	if err != nil {
		return nil, err
	}

	return &outside.UpdateStatusEvent{
		UUID:      protoUpdateEvent.Uuid,
		OrderUuid: protoUpdateEvent.OrderUuid,
		NewStatus: order.OrderStatus(protoUpdateEvent.NewStatus),
		UpdatedAt: protoUpdateEvent.CreatedAt.AsTime(),
	}, nil
}

func (l *UpdateListener) getRequestIdFromHeaders(headers []kafka.Header) string {
	id := ""
	for _, h := range headers {
		if h.Key == xrequestid.XREQUEST_ID_KEY {
			id = string(h.Value)
		}
	}

	return id
}

func (l *UpdateListener) writeToDLQ(ctx context.Context, msg kafka.Message, reason, message string) error {
	headers := make([]kafka.Header, len(msg.Headers), len(msg.Headers)+4)
	copy(headers, msg.Headers)

	headers = append(headers,
		kafka.Header{Key: "source_topic", Value: []byte(msg.Topic)},
		kafka.Header{Key: "reason", Value: []byte(reason)},
		kafka.Header{Key: "message", Value: []byte(message)},
		kafka.Header{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
	)

	dlqMsg := kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := l.dlqWriter.WriteMessages(writeCtx, dlqMsg); err != nil {
		l.logger.Error("failed to send message to DLQ",
			zap.Error(err),
			zap.String("reason", reason),
			zap.String("message", message),
		)
		return err
	}

	l.logger.Info("message sent to DLQ",
		zap.String("reason", reason),
		zap.String("message", message),
	)

	return nil
}
