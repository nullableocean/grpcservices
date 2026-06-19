package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.uber.org/zap"
)

func (r *OrderRepository) Save(ctx context.Context, order *model.Order, events ...model.Event) error {
	r.logger.Debug("start save order",
		zap.String("order_uuid", order.UUID),
		zap.String("order_type", string(order.Type)),
		zap.String("order_side", string(order.Side)),
		zap.String("market_uuid", string(order.MarketUUID)),
		zap.String("price", order.Price.String()),
		zap.String("user_uuid", order.UserUUID),
	)

	var lastError error
	for attempts := 0; attempts < r.retries; attempts++ {
		if attempts > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.backoff(attempts)):
			}
		}

		err := r.saveOrder(ctx, order, events...)
		if err == nil {
			return nil
		}

		lastError = err
		if !r.isRetryableError(err) {
			return err
		}
	}

	return lastError
}

func (r *OrderRepository) saveOrder(ctx context.Context, order *model.Order, events ...model.Event) error {
	tx, err := r.pgpool.Begin(ctx)
	if err != nil {
		return r.mapDBError(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	sideID, err := r.getSideID(ctx, tx, order.Side)
	if err != nil {
		return err
	}

	typeID, err := r.getTypeID(ctx, tx, order.Type)
	if err != nil {
		return err
	}

	statusID, err := r.getStatusID(ctx, tx, order.Status)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO orders (uuid, user_uuid, market_uuid, side_id, order_type_id, order_status_id, price, quantity, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id
    `
	var orderID int64
	err = tx.QueryRow(ctx, query,
		order.UUID,
		order.UserUUID,
		order.MarketUUID,
		sideID,
		typeID,
		statusID,
		order.Price,
		order.Quantity,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&orderID)
	if err != nil {
		return r.mapDBError(err, "failed to save order in db")
	}

	err = r.writeEventsToOutbox(ctx, tx, orderID, events)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return r.mapDBError(err, "failed to commit transaction")
	}

	r.logger.Debug("success save order", zap.String("order_uuid", order.UUID))

	return nil
}

func (r *OrderRepository) isRetryableError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001": // serialization_failure (deadlock)
			return true
		case "40P01": // deadlock_detected
			return true
		case "57P01", "57P02", "57P03": // admin shutdown, crash, cannot connect
			return true
		default:
			return false
		}
	}

	return false
}
