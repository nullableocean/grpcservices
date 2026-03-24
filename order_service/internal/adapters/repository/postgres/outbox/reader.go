package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRecord struct {
	EventUUID   string
	OrderUUID   string
	EventType   string
	Payload     json.RawMessage
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

type OutboxReader struct {
	pool *pgxpool.Pool
}

func NewOutboxReader(pool *pgxpool.Pool) *OutboxReader {
	return &OutboxReader{pool: pool}
}

func (r *OutboxReader) FetchUnprocessedTx(ctx context.Context, tx pgx.Tx, limit int) ([]*OutboxRecord, error) {
	const query = `
        SELECT uuid, order_uuid, event_type, payload, created_at, processed_at
        FROM outbox_orders_events
        WHERE processed_at IS NULL
        ORDER BY created_at
        LIMIT $1
        FOR UPDATE SKIP LOCKED
    `
	rows, err := tx.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var records []*OutboxRecord
	for rows.Next() {
		var rec OutboxRecord
		err := rows.Scan(
			&rec.EventUUID, &rec.OrderUUID, &rec.EventType, &rec.Payload,
			&rec.CreatedAt, &rec.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}
		records = append(records, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return records, nil
}

func (r *OutboxReader) MarkProcessedTx(ctx context.Context, tx pgx.Tx, uuid string) error {
	const query = `UPDATE outbox_orders_events SET processed_at = NOW() WHERE uuid = $1`
	_, err := tx.Exec(ctx, query, uuid)
	if err != nil {
		return fmt.Errorf("failed to mark event in outbox as processed: %w", err)
	}
	return nil
}
