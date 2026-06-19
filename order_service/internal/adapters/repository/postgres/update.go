package postgres

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

func (r *OrderRepository) Update(ctx context.Context, updatedOrder *model.Order, events ...model.Event) error {
	tx, err := r.pgpool.Begin(ctx)
	if err != nil {
		return r.mapDBError(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	var orderID int64
	err = tx.QueryRow(ctx, `SELECT id FROM orders WHERE uuid = $1`, updatedOrder.UUID).Scan(&orderID)
	if err != nil {
		return r.mapDBError(err, "failed to find order")
	}

	statusID, err := r.getStatusID(ctx, tx, updatedOrder.Status)
	if err != nil {
		return err
	}

	query := `
        UPDATE orders
        SET order_status_id = $1, updated_at = NOW()
        WHERE uuid = $2
    `
	_, err = tx.Exec(ctx, query, statusID, updatedOrder.UUID)
	if err != nil {
		return r.mapDBError(err, "failed to update order in db")
	}

	err = r.writeEventsToOutbox(ctx, tx, orderID, events)
	if err != nil {
		return r.mapDBError(err, "failed to write events")
	}

	if err := tx.Commit(ctx); err != nil {
		return r.mapDBError(err, "failed to commit transaction")
	}

	return nil
}
