package writer

import (
	"context"
	"time"

	marketseventsv1 "github.com/nullableocean/grpcservices/api/gen/events/markets/v1"
	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/events"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UpdateWriter struct {
	kwriter *kafka.Writer
	logger  *zap.Logger
}

func NewUpdateWriter(logger *zap.Logger, kw *kafka.Writer) *UpdateWriter {
	return &UpdateWriter{
		kwriter: kw,
		logger:  logger,
	}
}

func (w *UpdateWriter) Write(ctx context.Context, event *events.MarketUpdateEvent) error {
	ctx = context.WithoutCancel(ctx)

	reqId := w.getRequestId(ctx)

	ctx, span := otel.Tracer("markets_update_event_writer").Start(ctx, "write_update_event")
	defer span.End()
	span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqId))

	protoEvent := &marketseventsv1.MarketUpdated{
		MarketUuid: event.MarketUuid,
		UpdatedAt:  timestamppb.New(event.UpdateAt),
	}

	data, err := proto.Marshal(protoEvent)
	if err != nil {
		span.AddEvent("failed marshal to proto")
		w.logger.Error("failed to marshal market updated event", zap.Error(err))
		return err
	}

	headers := w.getHeaders(ctx, reqId)

	msg := kafka.Message{
		Key:     []byte(event.MarketUuid),
		Value:   data,
		Headers: headers,
		Time:    time.Now(),
	}

	if err := w.kwriter.WriteMessages(ctx, msg); err != nil {
		span.AddEvent("failed write event")
		w.logger.Error("failed to write market update event to Kafka",
			zap.Error(err),
			zap.String("market_uuid", event.MarketUuid))
		return err
	}

	w.logger.Info("market update event sent", zap.String("market_uuid", event.MarketUuid))

	return nil
}

func (w *UpdateWriter) getHeaders(ctx context.Context, xreqid string) []kafka.Header {
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

func (w *UpdateWriter) getRequestId(ctx context.Context) string {
	id := xrequestid.GetFromIncomingCtx(ctx)
	if id == "" {
		return xrequestid.NewXRequestId()
	}

	return id
}
