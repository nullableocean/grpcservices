ALTER TABLE outbox_orders_events DROP CONSTRAINT IF EXISTS fk_outbox_orders_events_order_id;
ALTER TABLE outbox_orders_events DROP COLUMN order_id;