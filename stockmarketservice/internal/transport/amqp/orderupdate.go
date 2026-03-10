package amqp

import (
	"context"

	ordereventsv1 "github.com/nullableocean/grpcservices/api/gen/events/order/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderUpdateWriter struct {
	kafkaWriter *kafka.Writer
	logger      *zap.Logger
}

func NewOrderUpdateWriter(l *zap.Logger, kw *kafka.Writer) *OrderUpdateWriter {
	return &OrderUpdateWriter{
		kafkaWriter: kw,
		logger:      l,
	}
}

func (w *OrderUpdateWriter) Write(ctx context.Context, event *domain.OrderUpdate) error {
	ctx = context.WithoutCancel(ctx)
	reqId := w.getRequestId(ctx)

	logger := w.logger.With(
		zap.String(xrequestid.XREQUEST_ID_KEY, reqId),
		zap.String("order_uuid", event.OrderUuid),
		zap.String("event_uuid", event.UUID),
	)

	data, err := w.marshalToBytes(event)
	if err != nil {
		logger.Error("failed to marshal event", zap.Error(err))
		return err
	}

	logger.Info("writing event for update order")

	traceCtx, span := otel.Tracer("stockmarket_update_writer").Start(ctx, "order_update_write")
	span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqId))
	defer span.End()

	headers := w.getHeaders(traceCtx, reqId)
	msg := kafka.Message{
		Key:     []byte(event.OrderUuid),
		Value:   data,
		Headers: headers,
		Time:    event.CreatedAt,
	}

	err = w.kafkaWriter.WriteMessages(traceCtx, msg)
	if err != nil {
		logger.Error("failed write event", zap.Error(err))
		span.AddEvent("failed write event")
		return err
	}

	logger.Info("success writed event")
	span.AddEvent("success write event")
	return nil
}

func (w *OrderUpdateWriter) marshalToBytes(event *domain.OrderUpdate) ([]byte, error) {
	protoEvent := &ordereventsv1.UpdateStatus{
		Uuid:      event.UUID,
		OrderUuid: event.OrderUuid,
		NewStatus: typesv1.OrderStatus(event.NewStatus),
		CreatedAt: timestamppb.New(event.CreatedAt),
	}

	b, err := proto.Marshal(protoEvent)
	return b, err
}

func (w *OrderUpdateWriter) getRequestId(ctx context.Context) string {
	id := xrequestid.GetFromIncomingCtx(ctx)
	if id == "" {
		return xrequestid.NewXRequestId()
	}

	return id
}

func (w *OrderUpdateWriter) getHeaders(ctx context.Context, xreqid string) []kafka.Header {
	carrier := propagation.HeaderCarrier{}

	headers := make([]kafka.Header, 0, len(carrier)+1)

	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, vals := range carrier {
		if len(vals) > 0 {
			headers = append(headers, kafka.Header{Key: k, Value: []byte(vals[0])})
		}
	}

	headers = append(headers, kafka.Header{Key: xrequestid.XREQUEST_ID_KEY, Value: []byte(xreqid)})

	return headers
}
