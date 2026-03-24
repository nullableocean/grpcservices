CREATE TABLE IF NOT EXISTS order_sides (
    id   SMALLINT PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE
);

INSERT INTO order_sides (id, code) VALUES
    (1, 'BUY'),
    (2, 'SELL');

CREATE TABLE IF NOT EXISTS order_types (
    id   SMALLINT PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE
);

INSERT INTO order_types (id, code) VALUES
    (1, 'LIMIT'),
    (2, 'MARKET'),
    (3, 'STOP_LOSS'),
    (4, 'TAKE_PROFIT');

CREATE TABLE IF NOT EXISTS order_statuses (
    id   SMALLINT PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE
);

INSERT INTO order_statuses (id, code) VALUES
    (1, 'CREATED'),
    (2, 'PENDING'),
    (3, 'COMPLETED'),
    (4, 'CANCELLED'),
    (5, 'REJECTED');

CREATE TABLE IF NOT EXISTS orders (
    id               BIGSERIAL PRIMARY KEY,
    uuid             UUID NOT NULL UNIQUE,
    user_uuid        UUID NOT NULL,
    market_uuid      UUID NOT NULL,
    side_id          SMALLINT NOT NULL REFERENCES order_sides(id),
    order_type_id    SMALLINT NOT NULL REFERENCES order_types(id),
    order_status_id  SMALLINT NOT NULL REFERENCES order_statuses(id),
    price            DECIMAL NOT NULL,
    quantity         DECIMAL NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_uuid ON orders (uuid);
CREATE INDEX idx_orders_user_uuid ON orders (user_uuid);

CREATE TABLE IF NOT EXISTS outbox_orders_events (
    id            BIGSERIAL PRIMARY KEY,
    uuid          UUID NOT NULL UNIQUE,
    order_uuid    UUID NOT NULL,
    event_type    VARCHAR(100) NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX idx_outbox_order_uuid ON outbox_orders_events (order_uuid);
CREATE INDEX idx_outbox_unprocessed ON outbox_orders_events (created_at) WHERE processed_at IS NULL;