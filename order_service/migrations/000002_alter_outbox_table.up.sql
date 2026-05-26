ALTER TABLE outbox_orders_events ADD COLUMN order_id BIGINT;

UPDATE outbox_orders_events
SET order_id = orders.id
FROM orders
WHERE outbox_orders_events.order_uuid = orders.uuid;

ALTER TABLE outbox_orders_events ALTER COLUMN order_id SET NOT NULL;

ALTER TABLE outbox_orders_events
    ADD CONSTRAINT fk_outbox_orders_events_order_id
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;