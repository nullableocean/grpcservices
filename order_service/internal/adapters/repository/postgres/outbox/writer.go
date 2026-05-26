package outbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type OutboxWriter struct{}

func NewOutboxWriter() *OutboxWriter {
	return &OutboxWriter{}
}

func (w *OutboxWriter) Write(ctx context.Context, tx pgx.Tx, orderRowId int64, event model.Event) error {
	payload, err := event.Payload()
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := `
        INSERT INTO outbox_orders_events (uuid, order_id, order_uuid, event_type, payload)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err = tx.Exec(ctx, query, event.ID(), orderRowId, event.GetOrderUUID(), event.EventType(), payload)
	if err != nil {
		return fmt.Errorf("failed to insert outbox record: %w", err)
	}

	return nil
}
