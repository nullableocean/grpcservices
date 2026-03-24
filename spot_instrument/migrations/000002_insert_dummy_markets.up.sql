-- DUMMY DEBUG MARKETS

INSERT INTO markets (uuid, name, is_enabled, deleted_at, created_at, updated_at)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'Admin Market', true, NULL, NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'Moder Market', true, NULL, NOW(), NOW()),
    ('33333333-3333-3333-3333-333333333333', 'Common Market', true, NULL, NOW(), NOW())
ON CONFLICT (uuid) DO NOTHING;

INSERT INTO market_allowed_roles (market_uuid, role_id)
VALUES
    ('11111111-1111-1111-1111-111111111111', 5),
    ('22222222-2222-2222-2222-222222222222', 4),
    ('33333333-3333-3333-3333-333333333333', 5),
    ('33333333-3333-3333-3333-333333333333', 4)
ON CONFLICT (market_uuid, role_id) DO NOTHING;