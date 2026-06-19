package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

func (r *OrderRepository) FindByUUID(ctx context.Context, orderUUID string) (*model.Order, error) {
	query := `
        SELECT o.uuid, o.user_uuid, o.market_uuid,
               s.code AS order_side,
               t.code AS order_type,
               st.code AS order_status,
               o.price, o.quantity, o.created_at, o.updated_at
        FROM orders o
        JOIN order_sides s ON o.side_id = s.id
        JOIN order_types t ON o.order_type_id = t.id
        JOIN order_statuses st ON o.order_status_id = st.id
        WHERE o.uuid = $1
    `
	var order model.Order
	var orderSide, orderType, orderStatus string
	err := r.pgpool.QueryRow(ctx, query, orderUUID).Scan(
		&order.UUID,
		&order.UserUUID,
		&order.MarketUUID,
		&orderSide,
		&orderType,
		&orderStatus,
		&order.Price,
		&order.Quantity,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, r.mapDBError(err, "order not found")
		}

		return nil, r.mapDBError(err, "failed to find order")
	}

	order.Side = model.OrderSide(orderSide)
	order.Type = model.OrderType(orderType)
	order.Status = model.OrderStatus(orderStatus)

	return &order, nil
}
