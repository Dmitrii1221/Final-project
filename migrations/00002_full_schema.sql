-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE currencies (
    id           BIGSERIAL    PRIMARY KEY,
    code         TEXT         NOT NULL UNIQUE,
    display_name TEXT         NOT NULL,
    precision    INT          NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id            BIGSERIAL  PRIMARY KEY,
    external_id   TEXT       UNIQUE,
    username      TEXT       UNIQUE,
    password_hash TEXT       NOT NULL,
    email         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT      NOT NULL UNIQUE,
    description TEXT
);

ALTER TABLE budgets
    ADD COLUMN owner_user_id BIGINT REFERENCES users(id);

CREATE TABLE budget_periods (
    id           BIGSERIAL    PRIMARY KEY,
    budget_id    BIGINT       NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    period_start TIMESTAMPTZ  NOT NULL,
    period_end   TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (period_end > period_start),
    EXCLUDE USING gist (
        budget_id WITH =,
        tstzrange(period_start, period_end, '[)') WITH &&
    )
);

CREATE TABLE period_limits (
    id           BIGSERIAL PRIMARY KEY,
    period_id    BIGINT    NOT NULL REFERENCES budget_periods(id) ON DELETE CASCADE,
    currency_id  BIGINT    NOT NULL REFERENCES currencies(id),
    limit_amount NUMERIC   NOT NULL,
    UNIQUE (period_id, currency_id)
);

CREATE TABLE period_balances (
    period_id    BIGINT       NOT NULL REFERENCES budget_periods(id) ON DELETE CASCADE,
    currency_id  BIGINT       NOT NULL REFERENCES currencies(id),
    remaining    NUMERIC      NOT NULL,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (period_id, currency_id)
);

CREATE TABLE spendings (
    id               BIGSERIAL    PRIMARY KEY,
    period_id        BIGINT       NOT NULL REFERENCES budget_periods(id) ON DELETE CASCADE,
    budget_id        BIGINT       NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    currency_id      BIGINT       NOT NULL REFERENCES currencies(id),
    amount           NUMERIC      NOT NULL,
    idempotency_key  TEXT,
    spent_at         TIMESTAMPTZ  NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (budget_id, idempotency_key)
);

CREATE TABLE user_budget_roles (
    user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    budget_id BIGINT NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    role_id   BIGINT NOT NULL REFERENCES roles(id),
    PRIMARY KEY (user_id, budget_id, role_id)
);

-- +goose Down
DROP TABLE user_budget_roles;
DROP TABLE spendings;
DROP TABLE period_balances;
DROP TABLE period_limits;
DROP TABLE budget_periods;

ALTER TABLE budgets DROP COLUMN owner_user_id;

DROP TABLE roles;
DROP TABLE users;
DROP TABLE currencies;

DROP EXTENSION IF EXISTS btree_gist;
