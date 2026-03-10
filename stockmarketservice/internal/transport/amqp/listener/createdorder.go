package listener

import (
	"context"
	"errors"
	"sync"
	"time"

	ordereventsv1 "github.com/nullableocean/grpcservices/api/gen/events/order/v1"
	"github.com/nullableocean/grpcservices/shared/limiter"
	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/errs"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/processor"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/transport/mapping"
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

type CreatedOrderListener struct {
	kafkaReader *kafka.Reader
	dlqWriter   *kafka.Writer

	processor *processor.StockmarketProcessor
	logger    *zap.Logger

	maxRetry      int
	handleRetries map[string]int
	mu            sync.Mutex

	processLimiter *limiter.Limiter
}

type Option struct {
	ProcessLimit int
	MaxRetries   int
}

func NewCreatedOrderListener(
	logger *zap.Logger,
	kreader *kafka.Reader,
	dlqWriter *kafka.Writer,
	processor *processor.StockmarketProcessor,
	opt Option,
) *CreatedOrderListener {
	limit := opt.ProcessLimit
	retries := opt.MaxRetries

	if limit <= 0 {
		limit = defaultProcLimit
	}
	if retries < 0 {
		retries = defaultRetries
	}

	return &CreatedOrderListener{
		kafkaReader:    kreader,
		dlqWriter:      dlqWriter,
		processor:      processor,
		logger:         logger,
		maxRetry:       retries,
		handleRetries:  make(map[string]int),
		processLimiter: limiter.New(limit),
	}
}

func (l *CreatedOrderListener) StartListen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("stop created order listener by context")
			return ctx.Err()
		default:
		}

		msg, err := l.kafkaReader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				l.logger.Info("listener stopped by context", zap.Error(err))
				return err
			}
			l.logger.Error("failed to fetch message from Kafka", zap.Error(err))
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := l.processLimiter.AcquireContext(ctx); err != nil {
			l.logger.Warn("context cancelled while acquiring limiter", zap.Error(err))
			return ctx.Err()
		}

		go l.handleMsg(ctx, msg)
	}
}

func (l *CreatedOrderListener) handleMsg(parentCtx context.Context, msg kafka.Message) {
	defer l.processLimiter.Release()

	traceCtx, span := l.startTracing(parentCtx, msg)
	defer span.End()

	msgKey := string(msg.Key)
	reqId := l.getRequestIdFromHeaders(msg.Headers)
	logger := l.logger.With(
		zap.String(xrequestid.XREQUEST_ID_KEY, reqId),
		zap.String("msg_key", msgKey),
	)

	logger.Info("read created order event from kafka", zap.String("topice", l.kafkaReader.Config().Topic))
	span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqId))

	event, err := l.unmarshalEvent(msg.Value)
	if err != nil {
		logger.Error("failed to unmarshal event", zap.Error(err))
		span.AddEvent("unmarshal_error")

		l.kafkaReader.CommitMessages(traceCtx, msg)
		l.writeToDLQ(traceCtx, msg, "unmarshal_error", err.Error())
		return
	}

	order := mapping.MapProtoOrderToDomainOrder(event.CreatedOrder)

	logger.Info("start process orde from kafka event")
	err = l.processor.Process(traceCtx, order)
	if err != nil {
		logger.Error("failed to process event order", zap.Error(err), zap.String("event_uuid", event.EventUuid))

		if !errors.Is(err, errs.ErrAlreadyProcessed) && !errors.Is(err, errs.ErrAlreadyProcessing) {
			l.mu.Lock()
			l.handleRetries[event.EventUuid] += 1

			if l.handleRetries[event.EventUuid] >= l.maxRetry {
				logger.Warn("max retries exceeded, sending to DLQ", zap.String("event_uuid", event.EventUuid))
				span.AddEvent("max_retries_exceeded")

				l.kafkaReader.CommitMessages(traceCtx, msg)
				l.writeToDLQ(traceCtx, msg, "max_retries_exceeded", "")

				delete(l.handleRetries, event.EventUuid)
				l.mu.Unlock()
				return
			}
			l.mu.Unlock()

			span.AddEvent("event_retry_scheduled")
			return
		}
	}

	if err := l.kafkaReader.CommitMessages(traceCtx, msg); err != nil {
		logger.Error("failed to commit offset after successful handling", zap.Error(err))
		span.AddEvent("commit_error")
		return
	}

	l.mu.Lock()
	delete(l.handleRetries, event.EventUuid)
	l.mu.Unlock()

	logger.Info("successfully handled created order event", zap.String("event_uuid", event.EventUuid))
	span.AddEvent("commit_success")
}

func (l *CreatedOrderListener) startTracing(ctx context.Context, msg kafka.Message) (context.Context, trace.Span) {
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.HeaderCarrier{}
	for _, h := range msg.Headers {
		carrier.Set(h.Key, string(h.Value))
	}

	ctx = propagator.Extract(ctx, carrier)
	traceCtx, span := otel.Tracer("stockmarket_created_order_listener").Start(ctx, "handle_created_order")
	return traceCtx, span
}

func (l *CreatedOrderListener) unmarshalEvent(data []byte) (*ordereventsv1.CreatedOrderEvent, error) {
	event := &ordereventsv1.CreatedOrderEvent{}
	err := proto.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (l *CreatedOrderListener) getRequestIdFromHeaders(headers []kafka.Header) string {
	for _, h := range headers {
		if h.Key == xrequestid.XREQUEST_ID_KEY {
			return string(h.Value)
		}
	}
	return ""
}

func (l *CreatedOrderListener) writeToDLQ(ctx context.Context, msg kafka.Message, reason, message string) error {
	headers := make([]kafka.Header, len(msg.Headers), len(msg.Headers)+5)
	copy(headers, msg.Headers)

	headers = append(headers,
		kafka.Header{Key: "source_topic", Value: []byte(msg.Topic)},
		kafka.Header{Key: "reason", Value: []byte(reason)},
		kafka.Header{Key: "message", Value: []byte(message)},
		kafka.Header{Key: "original_timestamp", Value: []byte(msg.Time.Format(time.RFC3339))},
		kafka.Header{Key: "dlq_timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
	)

	dlqMsg := kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := l.dlqWriter.WriteMessages(writeCtx, dlqMsg); err != nil {
		l.logger.Error("failed to write event to DLQ",
			zap.Error(err),
			zap.String("reason", reason),
			zap.String("message", message),
		)
		return err
	}

	l.logger.Info("event sent to DLQ",
		zap.String("reason", reason),
		zap.String("message", message),
	)
	return nil
}
