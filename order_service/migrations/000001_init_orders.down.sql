DROP INDEX IF EXISTS idx_outbox_pending_events;
DROP INDEX IF EXISTS idx_outbox_order_uuid;
DROP INDEX IF EXISTS idx_orders_user_uuid;
DROP INDEX IF EXISTS idx_orders_uuid;

DROP TABLE IF EXISTS outbox_orders_events;
DROP TYPE IF EXISTS outbox_event_status;

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS order_sides;
DROP TABLE IF EXISTS order_types;
DROP TABLE IF EXISTS order_statuses;

