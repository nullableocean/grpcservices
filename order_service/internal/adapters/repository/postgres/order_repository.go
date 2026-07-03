package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
)

var _ ports.OrderRepository = &OrderRepository{}

type OrderRepository struct {
	pgpool *pgxpool.Pool
	outbox *outbox.OutboxWriter
	logger *logger.CtxZapLogger

	retries int
	backoff func(attempt int) time.Duration
}

type RepositoryConfig struct {
	Retries          int
	RetryBackoffFunc func(attempt int) time.Duration
}

func (cfg RepositoryConfig) Validate() error {
	if cfg.Retries < 0 {
		return fmt.Errorf("%w: negative repository retries", errs.ErrIncorrectData)
	}

	if cfg.RetryBackoffFunc == nil {
		return fmt.Errorf("%w: retry backoff callback is nil", errs.ErrIncorrectData)
	}

	return nil
}

func NewOrderRepository(l *logger.CtxZapLogger, pool *pgxpool.Pool, outbox *outbox.OutboxWriter, cfg RepositoryConfig) (*OrderRepository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &OrderRepository{
		pgpool:  pool,
		outbox:  outbox,
		logger:  l,
		retries: cfg.Retries,
		backoff: cfg.RetryBackoffFunc,
	}, nil
}

func (r *OrderRepository) writeEventsToOutbox(ctx context.Context, tx pgx.Tx, orderRowId int64, events []model.Event) error {
	for _, event := range events {
		if err := r.outbox.Write(ctx, tx, orderRowId, event); err != nil {
			return r.mapDBError(err, "failed to write event in outbox")
		}
	}

	return nil
}

func (r *OrderRepository) getSideID(ctx context.Context, tx pgx.Tx, side model.OrderSide) (int, error) {
	var id int

	err := tx.QueryRow(ctx, `SELECT id FROM order_sides WHERE code = $1`, string(side)).Scan(&id)
	if err != nil {
		return 0, r.mapDBError(err, "failed to get order side id")
	}

	return id, nil
}

func (r *OrderRepository) getTypeID(ctx context.Context, tx pgx.Tx, orderType model.OrderType) (int, error) {
	var id int

	err := tx.QueryRow(ctx, `SELECT id FROM order_types WHERE code = $1`, string(orderType)).Scan(&id)
	if err != nil {
		return 0, r.mapDBError(err, "failed to get order type id")
	}

	return id, nil
}

func (r *OrderRepository) getStatusID(ctx context.Context, tx pgx.Tx, status model.OrderStatus) (int, error) {
	var id int

	err := tx.QueryRow(ctx, `SELECT id FROM order_statuses WHERE code = $1`, string(status)).Scan(&id)
	if err != nil {
		return 0, r.mapDBError(err, "failed to get order status id")
	}

	return id, nil
}

func (r *OrderRepository) getStatusIDWithoutTx(ctx context.Context, status model.OrderStatus) (int, error) {
	var id int
	err := r.pgpool.QueryRow(ctx, `SELECT id FROM order_statuses WHERE code = $1`, string(status)).Scan(&id)
	if err != nil {
		return 0, r.mapDBError(err, "failed to get order status id")
	}
	return id, nil
}

func (r *OrderRepository) getTypeIDWithoutTx(ctx context.Context, orderType model.OrderType) (int, error) {
	var id int
	err := r.pgpool.QueryRow(ctx, `SELECT id FROM order_types WHERE code = $1`, string(orderType)).Scan(&id)
	if err != nil {
		return 0, r.mapDBError(err, "failed to get order type id")
	}
	return id, nil
}

func (r *OrderRepository) mapDBError(err error, description string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", description, errs.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%s: %w", description, errs.ErrDuplicateKey)
		case "23503":
			return fmt.Errorf("%s: %w", description, errs.ErrForeignKeyViolation)
		case "23514":
			return fmt.Errorf("%s: %w", description, errs.ErrInvalidInput)
		}
	}

	return fmt.Errorf("%s: %w: %w", description, errs.ErrInternal, err)
}
