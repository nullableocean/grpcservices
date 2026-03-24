CREATE TABLE IF NOT EXISTS roles (
    id   SMALLINT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL
);

INSERT INTO roles (id, code, name) VALUES
    (1, 'GUEST', 'Guest'),
    (2, 'TRADER', 'Trader'),
    (3, 'MARKET_MAKER', 'Market Maker'),
    (4, 'MODER', 'Moderator'),
    (5, 'ADMIN', 'Administrator');

CREATE TABLE IF NOT EXISTS markets (
    uuid        UUID PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    is_enabled  BOOLEAN NOT NULL DEFAULT true,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_markets_active ON markets (is_enabled, deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS market_allowed_roles (
    market_uuid UUID NOT NULL REFERENCES markets(uuid) ON DELETE CASCADE,
    role_id     SMALLINT NOT NULL REFERENCES roles(id),
    PRIMARY KEY (market_uuid, role_id)
);

CREATE INDEX idx_market_allowed_roles_role_id ON market_allowed_roles (role_id);