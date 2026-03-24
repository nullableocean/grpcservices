package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

var (
	defaultInterval  = 5 * time.Second
	defaultBatchSize = 20
)

type OutboxRelay struct {
	publisher ports.EventPublisher

	reader *OutboxReader
	pgpool *pgxpool.Pool

	interval  time.Duration
	batchSize int

	logger *zap.Logger
}

type Options struct {
	interval  time.Duration
	batchSize int
}

func NewRelay(l *zap.Logger, pool *pgxpool.Pool, publisher ports.EventPublisher, opt Options) *OutboxRelay {
	if opt.interval <= 0 {
		opt.interval = defaultInterval
	}
	if opt.batchSize <= 0 {
		opt.batchSize = defaultBatchSize
	}

	return &OutboxRelay{
		reader:    NewOutboxReader(pool),
		publisher: publisher,
		pgpool:    pool,
		interval:  opt.interval,
		batchSize: opt.batchSize,
		logger:    l,
	}
}

func (r *OutboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("event outbox relay closed by context")
			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *OutboxRelay) processBatch(ctx context.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("panic in outbox relay", zap.Any("error", rec), zap.Stack("stacktrace"))
			return
		}
	}()

	tx, err := r.pgpool.Begin(ctx)
	if err != nil {
		r.logger.Error("failed to begin transaction in relay", zap.Error(err))
		return
	}
	defer tx.Rollback(ctx)

	records, err := r.reader.FetchUnprocessedTx(ctx, tx, r.batchSize)
	if err != nil {
		r.logger.Error("failed fetch unprocessed events", zap.Error(err))
		return
	}

	if len(records) == 0 {
		tx.Commit(ctx)
		return
	}

	for _, rec := range records {
		event, err := r.unmarshalEvent(rec)
		if err != nil {
			r.logger.Error("failed unmarshaling outbox record to event", zap.Error(err), zap.String("event_uuid", rec.EventUUID))
			continue
		}

		if err := r.publisher.Publish(ctx, event); err != nil {
			r.logger.Error("failed publish event", zap.Error(err), zap.String("event_uuid", rec.EventUUID))
			continue
		}

		if err := r.reader.MarkProcessedTx(ctx, tx, rec.EventUUID); err != nil {
			r.logger.Error("failed mark event as processed in outbox", zap.Error(err), zap.String("event_uuid", rec.EventUUID))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("failed commit relay transaction")
	}
}

func (r *OutboxRelay) unmarshalEvent(rec *OutboxRecord) (model.Event, error) {
	switch model.EventType(rec.EventType) {
	case model.EVENT_ORDER_CREATED:
		data := &model.EventCreatedData{}
		if err := json.Unmarshal(rec.Payload, data); err != nil {
			return nil, fmt.Errorf("unmarshal order.created event payload: %w", err)
		}
		return &model.EventOrderCreated{
			OrderUUID: rec.OrderUUID,
			Data:      data,
		}, nil

	case model.EVENT_ORDER_UPDATED:
		data := &model.EventUpdatedData{}
		if err := json.Unmarshal(rec.Payload, data); err != nil {
			return nil, fmt.Errorf("unmarshal order.updated event payload: %w", err)
		}

		return &model.EventOrderUpdated{
			OrderUUID: rec.OrderUUID,
			Data:      data,
		}, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", rec.EventType)
	}
}
