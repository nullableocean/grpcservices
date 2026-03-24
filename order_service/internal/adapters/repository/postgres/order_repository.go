package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

var _ ports.OrderRepository = &OrderRepository{}

type OrderRepository struct {
	pgpool *pgxpool.Pool
	outbox *outbox.OutboxWriter
	logger *zap.Logger
}

func NewOrderRepository(l *zap.Logger, pool *pgxpool.Pool, outbox *outbox.OutboxWriter) *OrderRepository {
	return &OrderRepository{
		pgpool: pool,
		outbox: outbox,
		logger: l,
	}
}

func (r *OrderRepository) getSideID(ctx context.Context, tx pgx.Tx, side model.OrderSide) (int, error) {
	var id int

	err := tx.QueryRow(ctx, `SELECT id FROM order_sides WHERE code = $1`, string(side)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get order side id: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) getTypeID(ctx context.Context, tx pgx.Tx, orderType model.OrderType) (int, error) {
	var id int

	err := tx.QueryRow(ctx, `SELECT id FROM order_types WHERE code = $1`, string(orderType)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get order type id: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) getStatusID(ctx context.Context, tx pgx.Tx, status model.OrderStatus) (int, error) {
	var id int

	err := tx.QueryRow(ctx, `SELECT id FROM order_statuses WHERE code = $1`, string(status)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get order status id: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) Save(ctx context.Context, order *model.Order, events ...model.Event) error {
	r.logger.Info("start save order",
		zap.String("order_uuid", order.UUID),
		zap.String("order_type", string(order.Type)),
		zap.String("order_side", string(order.Side)),
		zap.String("market_uuid", string(order.MarketUUID)),
		zap.String("price", order.Price.String()),
		zap.String("user_uuid", order.UserUUID),
	)

	tx, err := r.pgpool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
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
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
	_, err = tx.Exec(ctx, query,
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
	)
	if err != nil {
		return fmt.Errorf("failed to save order in db: %w", err)
	}

	err = r.writeEvents(ctx, tx, events)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("success save order", zap.String("order_uuid", order.UUID))

	return nil
}

func (r *OrderRepository) Update(ctx context.Context, updatedOrder *model.Order, events ...model.Event) error {
	tx, err := r.pgpool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE uuid = $1)`, updatedOrder.UUID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed query exist order: %w", err)
	}
	if !exists {
		return fmt.Errorf("order not found in db: %w", errs.ErrNotFound)
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
		return fmt.Errorf("failed update order in db: %w", err)
	}

	err = r.writeEvents(ctx, tx, events)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *OrderRepository) writeEvents(ctx context.Context, tx pgx.Tx, events []model.Event) error {
	for _, event := range events {
		if err := r.outbox.Write(ctx, tx, event); err != nil {
			return fmt.Errorf("failed to write event in outbox: %w", err)
		}
	}

	return nil
}

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
			return nil, fmt.Errorf("order not found: %w", err)
		}

		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	order.Side = model.OrderSide(orderSide)
	order.Type = model.OrderType(orderType)
	order.Status = model.OrderStatus(orderStatus)

	return &order, nil
}
