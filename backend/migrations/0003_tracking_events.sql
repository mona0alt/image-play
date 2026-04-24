-- Migration 0003_tracking_events: create tracking_events table

CREATE TABLE IF NOT EXISTS tracking_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    event VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
