-- Migration 0001_init: create core business tables
-- Tables: users, assets, generations, orders, transactions, system_configs

-- users
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) UNIQUE NOT NULL,
    balance NUMERIC(10,2) DEFAULT 0,
    free_quota INT DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- assets
CREATE TABLE IF NOT EXISTS assets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    object_key VARCHAR(255) NOT NULL,
    url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- generations
CREATE TABLE IF NOT EXISTS generations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    client_request_id VARCHAR(64) NOT NULL,
    scene_key VARCHAR(32),
    template_key VARCHAR(64),
    fields JSONB,
    source_asset_id BIGINT REFERENCES assets(id),
    status VARCHAR(32) NOT NULL,
    result_url VARCHAR(500),
    prompt TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- orders
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    order_no VARCHAR(64) UNIQUE NOT NULL,
    package_code VARCHAR(32),
    amount NUMERIC(10,2),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    wx_prepay_id VARCHAR(128),
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- transactions
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    generation_id BIGINT REFERENCES generations(id),
    type VARCHAR(32) NOT NULL,
    amount NUMERIC(10,2),
    balance_before NUMERIC(10,2),
    balance_after NUMERIC(10,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- system_configs
CREATE TABLE IF NOT EXISTS system_configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(64) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Idempotency and concurrency constraints for generations
CREATE UNIQUE INDEX IF NOT EXISTS idx_generations_user_request
ON generations(user_id, client_request_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_generations_user_active
ON generations(user_id)
WHERE status IN ('queued', 'running', 'result_auditing');

-- Indexes on foreign-key columns
CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_generations_user_id ON generations(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_generation_id ON transactions(generation_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_generation_id_unique ON transactions(generation_id);
