DROP INDEX IF EXISTS idx_outbox_unprocessed;
DROP INDEX IF EXISTS idx_outbox_order_uuid;
DROP INDEX IF EXISTS idx_orders_user_uuid;
DROP INDEX IF EXISTS idx_orders_uuid;

DROP TABLE IF EXISTS outbox_orders_events;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS order_sides;
DROP TABLE IF EXISTS order_types;
DROP TABLE IF EXISTS order_statuses;

