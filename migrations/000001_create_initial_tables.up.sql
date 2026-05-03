CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    login         VARCHAR(256) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_login ON users(login);


CREATE TABLE IF NOT EXISTS orders (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    number      VARCHAR(256) NOT NULL UNIQUE,
    user_id     UUID NOT NULL REFERENCES users(id),
    status      VARCHAR(64) NOT NULL DEFAULT 'NEW',
    accrual     DECIMAL(12, 2) DEFAULT 0,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_number ON orders(number);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);


CREATE TABLE IF NOT EXISTS balances (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id       UUID NOT NULL REFERENCES users(id),
    order_number  VARCHAR(256) NOT NULL UNIQUE,
    sum           DECIMAL(12, 2) NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);