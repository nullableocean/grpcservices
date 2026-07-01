package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type OutboxRecord struct {
	EventUUID string
	OrderUUID string
	EventType string
	Attempts  int
	Payload   json.RawMessage
	Error     string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type OutboxReader struct {
	pool *pgxpool.Pool
}

func NewOutboxReader(pool *pgxpool.Pool) *OutboxReader {
	return &OutboxReader{pool: pool}
}

func (r *OutboxReader) FetchUnprocessedTx(ctx context.Context, tx pgx.Tx, limit int) ([]*OutboxRecord, error) {
	const query = `
        SELECT uuid, order_uuid, event_type, payload, attempts, error, created_at, updated_at
        FROM outbox_orders_events
        WHERE status IN ($1,$2)
        ORDER BY created_at
        LIMIT $3
        FOR UPDATE SKIP LOCKED
    `
	rows, err := tx.Query(ctx, query, model.OutboxStatusPending, model.OutboxStatusFailed, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var records []*OutboxRecord
	for rows.Next() {
		var rec OutboxRecord
		var errMsg sql.NullString
		err := rows.Scan(
			&rec.EventUUID, &rec.OrderUUID, &rec.EventType, &rec.Payload, &rec.Attempts, &errMsg,
			&rec.CreatedAt, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}

		if errMsg.Valid {
			rec.Error = errMsg.String
		}

		records = append(records, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return records, nil
}

func (r *OutboxReader) MarkProcessedTx(ctx context.Context, tx pgx.Tx, eventUUID string) error {
	const query = `UPDATE outbox_orders_events 
         SET status = $1, updated_at = NOW(), error = NULL, attempts = attempts + 1 
         WHERE uuid = $2`

	_, err := tx.Exec(ctx, query, model.OutboxStatusPending, eventUUID)
	if err != nil {
		return fmt.Errorf("failed to mark event in outbox as processed: %w", err)
	}

	return nil
}

func (r *OutboxReader) MarkFailedTx(ctx context.Context, tx pgx.Tx, eventUUID string, errorMessage string) error {
	const query = `UPDATE outbox_orders_events 
         SET status = $1, updated_at = NOW(), error = $2, attempts = attempts + 1 
         WHERE uuid = $3`

	_, err := tx.Exec(ctx, query, model.OutboxStatusFailed, errorMessage, eventUUID)
	if err != nil {
		return fmt.Errorf("failed to mark event in outbox as failed: %w", err)
	}

	return nil
}

func (r *OutboxReader) MarkDeadTx(ctx context.Context, tx pgx.Tx, eventUUID string, errorMessage string) error {
	const query = `UPDATE outbox_orders_events 
         SET status = $1, updated_at = NOW(), error = $2, attempts = attempts + 1 
         WHERE uuid = $3`

	_, err := tx.Exec(ctx, query, model.OutboxStatusDeadLetter, errorMessage, eventUUID)

	if err != nil {
		return fmt.Errorf("failed to mark event in outbox as dead: %w", err)
	}

	return nil
}
