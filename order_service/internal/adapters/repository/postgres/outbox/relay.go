package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

type OutboxRelay struct {
	publisher ports.EventPublisher
	reader    *OutboxReader
	pgpool    *pgxpool.Pool

	interval  time.Duration
	timeout   time.Duration
	batchSize int
	maxRetry  int

	cancel    context.CancelFunc
	stopped   atomic.Bool
	inProcess atomic.Bool

	metrics *metrics.OutboxMetricsRecorder
	logger  *zap.Logger
}

type Config struct {
	Interval     time.Duration
	BatchTimeout time.Duration
	BatchSize    int
	MaxRetry     int
}

func (c Config) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("invalid outbox relay interval. cant be negative or zero: %s", c.Interval)
	}

	if c.BatchTimeout <= 0 {
		return fmt.Errorf("invalid outbox batch timeout. cant be negative or zero: %s", c.BatchTimeout)
	}

	if c.BatchSize <= 0 {
		return fmt.Errorf("invalid outbox relay batch size. cant be negative or zero: %d", c.BatchSize)
	}

	if c.MaxRetry <= 0 {
		return fmt.Errorf("invalid outbox relay retry. cant be negative or zero: %d", c.MaxRetry)
	}

	return nil
}

func NewRelay(l *zap.Logger, pool *pgxpool.Pool, publisher ports.EventPublisher, outboxMetrics *metrics.OutboxMetricsRecorder, opt Config) (*OutboxRelay, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &OutboxRelay{
		reader:    NewOutboxReader(pool),
		publisher: publisher,
		pgpool:    pool,
		interval:  opt.Interval,
		batchSize: opt.BatchSize,
		timeout:   opt.BatchTimeout,
		maxRetry:  opt.MaxRetry,
		logger:    l,
		metrics:   outboxMetrics,
	}, nil
}

func (r *OutboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	for {
		select {
		case <-ctx.Done():
			r.logger.Debug("event outbox relay closed by context")
			return
		case <-ticker.C:
			if r.stopped.Load() {
				return
			}

			if r.inProcess.Load() {
				continue
			}

			r.inProcess.Store(true)
			go func() {
				defer r.inProcess.Store(false)
				r.processBatch(ctx)
			}()
		}
	}
}
func (r *OutboxRelay) processBatch(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	tx, err := r.pgpool.Begin(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			r.logger.Warn("timeout beginning transaction")
		} else {
			r.logger.Error("failed to begin transaction", zap.Error(err))
		}
		return
	}

	defer func() {
		err := tx.Rollback(context.WithoutCancel(ctx))
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("failed outbox transaction rollback", zap.Error(err))
		}
	}()

	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("panic in outbox relay", zap.Any("error", rec), zap.Stack("stacktrace"))
		}
	}()

	records, err := r.reader.FetchUnprocessedTx(ctx, tx, r.batchSize)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			r.logger.Warn("timeout fetching unprocessed events")
		} else {
			r.logger.Error("failed to fetch unprocessed events", zap.Error(err))
		}
		return
	}

	if len(records) == 0 {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			r.logger.Error("failed to commit empty transaction", zap.Error(commitErr))
		}
		return
	}

LOOP:
	for i, rec := range records {
		select {
		case <-ctx.Done():
			r.logger.Warn("processing cancelled by timeout or context",
				zap.Int("processed", i),
				zap.Int("total", len(records)),
				zap.Error(ctx.Err()),
			)
			break LOOP
		default:
		}
		r.metrics.EventFetched(ctx)

		if rec.Attempts >= r.maxRetry {
			markErr := r.reader.MarkDeadTx(ctx, tx, rec.EventUUID, rec.Error)
			if markErr != nil {
				r.logger.Error("failed mark dead event", zap.Error(err))
				return
			}
			continue
		}

		event, err := r.unmarshalEvent(rec)
		if err != nil {
			r.logger.Error("failed to unmarshal event",
				zap.Error(err),
				zap.String("event_uuid", rec.EventUUID))

			markErr := r.reader.MarkDeadTx(ctx, tx, rec.EventUUID, err.Error())
			if markErr != nil {
				r.logger.Error("failed mark dead event", zap.Error(err))
				return
			}

			r.metrics.EventFailed(ctx)
			continue
		}

		if err := r.publisher.Publish(ctx, event); err != nil {
			r.logger.Error("failed to publish event",
				zap.Error(err),
				zap.String("event_uuid", rec.EventUUID))

			markErr := r.reader.MarkFailedTx(ctx, tx, rec.EventUUID, err.Error())
			if markErr != nil {
				r.logger.Error("failed mark failed event", zap.Error(err))
				return
			}

			r.metrics.EventFailed(ctx)
			continue
		}

		if err := r.reader.MarkProcessedTx(ctx, tx, rec.EventUUID); err != nil {
			r.logger.Error("failed to mark event as processed",
				zap.Error(err),
				zap.String("event_uuid", rec.EventUUID))

			r.metrics.EventFailed(ctx)
			return
		}
	}

	if err := tx.Commit(context.WithoutCancel(ctx)); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return
	}

	r.metrics.EventRelayed(ctx)
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

func (r *OutboxRelay) Stop() error {
	if r.stopped.Load() {
		return fmt.Errorf("already stopped")
	}

	r.stopped.Store(true)
	if r.inProcess.Load() {
		<-time.After(r.timeout)
	}

	r.cancel()
	return nil
}
