package writer

import (
	"context"
	"time"

	"github.com/google/uuid"
	ordereventsv1 "github.com/nullableocean/grpcservices/api/gen/events/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/mapping"
	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type CreatedEventWriter struct {
	kwriter *kafka.Writer
	logger  *zap.Logger
}

func NewCreatedEventWriter(logger *zap.Logger, kw *kafka.Writer) *CreatedEventWriter {
	return &CreatedEventWriter{
		kwriter: kw,
		logger:  logger,
	}
}

func (w *CreatedEventWriter) Write(ctx context.Context, insideEvent *inside.OrderCreatedEvent) error {
	orderUuid := insideEvent.Order.UUID
	reqId := w.getRequestId(ctx)

	ctx, span := otel.Tracer("order_event_writer").Start(ctx, "write_created_event")
	defer span.End()

	span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqId))
	logger := w.logger.With(
		zap.String("order_uuid", orderUuid),
		zap.String(xrequestid.XREQUEST_ID_KEY, reqId),
	)

	protoEvent := &ordereventsv1.CreatedOrderEvent{
		EventUuid:    uuid.NewString(),
		CreatedOrder: mapping.MapDomainOrderToProtoOrder(insideEvent.Order),
	}

	data, err := proto.Marshal(protoEvent)
	if err != nil {
		logger.Error("failed to marshal created order event", zap.Error(err))
		return err
	}

	headers := w.prepareHeaders(ctx, reqId)
	msg := kafka.Message{
		Key:     []byte(insideEvent.Order.UUID),
		Value:   data,
		Headers: headers,
		Time:    insideEvent.Order.CreatedAt,
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logger.Info("write created order event to kafka", zap.String("topic", w.kwriter.Topic))

	if err := w.kwriter.WriteMessages(writeCtx, msg); err != nil {
		w.logger.Error("failed to write message to Kafka", zap.Error(err))
		return err
	}

	logger.Info("writed event to kafka", zap.String("event_uuid", protoEvent.EventUuid))
	return nil
}

func (w *CreatedEventWriter) prepareHeaders(ctx context.Context, requestId string) []kafka.Header {
	var headers []kafka.Header

	headers = append(headers, kafka.Header{
		Key:   xrequestid.XREQUEST_ID_KEY,
		Value: []byte(requestId),
	})

	carrier := propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for key, vals := range carrier {
		if len(vals) > 0 {
			headers = append(headers, kafka.Header{
				Key:   key,
				Value: []byte(vals[0]),
			})
		}
	}

	return headers
}

func (w *CreatedEventWriter) getRequestId(ctx context.Context) string {
	id := xrequestid.GetFromIncomingCtx(ctx)
	if id == "" {
		return xrequestid.NewXRequestId()
	}

	return id
}
