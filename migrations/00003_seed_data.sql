-- +goose Up
-- Сиды: валюты по умолчанию и роли

-- Валюты
INSERT INTO currencies (code, display_name, precision) VALUES
    ('USD', 'US Dollar', 2),
    ('TOKENS', 'LLM Tokens', 0)
ON CONFLICT (code) DO NOTHING;

-- Роли
INSERT INTO roles (code, description) VALUES
    ('owner', 'Создатель бюджета, полный доступ'),
    ('spender', 'Может тратить средства бюджета'),
    ('viewer', 'Может просматривать статистику и остатки'),
    ('limit_manager', 'Может изменять лимиты бюджета')
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DELETE FROM currencies WHERE code IN ('USD', 'TOKENS');
DELETE FROM roles WHERE code IN ('owner', 'spender', 'viewer', 'limit_manager');