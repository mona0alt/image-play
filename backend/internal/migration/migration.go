package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		openid VARCHAR(64) UNIQUE NOT NULL,
		balance NUMERIC(10,2) DEFAULT 0,
		free_quota INT DEFAULT 3,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	`CREATE TABLE IF NOT EXISTS assets (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT REFERENCES users(id),
		object_key VARCHAR(255) NOT NULL,
		url VARCHAR(500),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	`CREATE TABLE IF NOT EXISTS generations (
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
	);`,

	`CREATE TABLE IF NOT EXISTS orders (
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
	);`,

	`CREATE TABLE IF NOT EXISTS transactions (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT REFERENCES users(id),
		generation_id BIGINT REFERENCES generations(id),
		type VARCHAR(32) NOT NULL,
		amount NUMERIC(10,2),
		balance_before NUMERIC(10,2),
		balance_after NUMERIC(10,2),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	`CREATE TABLE IF NOT EXISTS system_configs (
		id BIGSERIAL PRIMARY KEY,
		config_key VARCHAR(64) UNIQUE NOT NULL,
		value JSONB NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	`CREATE TABLE IF NOT EXISTS scene_templates (
		id BIGSERIAL PRIMARY KEY,
		scene_key VARCHAR(32) NOT NULL,
		template_key VARCHAR(64) NOT NULL,
		name VARCHAR(128) NOT NULL,
		form_schema JSONB NOT NULL,
		prompt_preset JSONB NOT NULL,
		sample_image_url VARCHAR(255),
		is_active BOOLEAN DEFAULT TRUE,
		UNIQUE (scene_key, template_key)
	);`,

	`CREATE TABLE IF NOT EXISTS tracking_events (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		event VARCHAR(64) NOT NULL,
		payload JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	// Indexes
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_generations_user_request ON generations(user_id, client_request_id);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_generations_user_active ON generations(user_id) WHERE status IN ('queued', 'running', 'result_auditing');`,
	`CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_generations_user_id ON generations(user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_transactions_generation_id ON transactions(generation_id);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_generation_id_unique ON transactions(generation_id);`,
	`CREATE INDEX IF NOT EXISTS idx_tracking_events_created_at ON tracking_events(created_at);`,

	// Explore feed support
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS nickname VARCHAR(64) DEFAULT '';`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500) DEFAULT '';`,
	`CREATE TABLE IF NOT EXISTS likes (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id),
		generation_id BIGINT NOT NULL REFERENCES generations(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (user_id, generation_id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_likes_generation_id ON likes(generation_id);`,
	`CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);`,
}

var seedSQL = []string{
	`INSERT INTO scene_templates (scene_key, template_key, name, form_schema, prompt_preset, sample_image_url)
	VALUES
	  ('portrait', 'office-pro', '通勤职业', '[{"name":"subject_name","label":"拍摄对象","type":"text","required":true}]', '{"base_prompt":"职业形象照","style_words":["professional","business"]}', 'https://example.com/portrait-office-pro.png'),
	  ('festival', 'spring-festival', '春节祝福', '[{"name":"title","label":"标题","type":"text","required":true}]', '{"base_prompt":"春节贺卡","style_words":["festive","red and gold"]}', 'https://example.com/festival-spring.png'),
	  ('invitation', 'wedding-classic', '婚礼请柬', '[{"name":"host_name","label":"主办人","type":"text","required":true}]', '{"base_prompt":"婚礼请柬","style_words":["elegant","romantic"]}', 'https://example.com/invitation-wedding.png'),
	  ('tshirt', 'streetwear', '街头潮流', '[{"name":"theme","label":"主题","type":"text","required":true}]', '{"base_prompt":"街头潮流T恤图案","style_words":["streetwear","graffiti"]}', 'https://example.com/tshirt-streetwear.png'),
	  ('poster', 'concert', '演唱会海报', '[{"name":"title","label":"标题","type":"text","required":true}]', '{"base_prompt":"演唱会海报","style_words":["concert","neon"]}', 'https://example.com/poster-concert.png')
	ON CONFLICT (scene_key, template_key) DO UPDATE
	SET name = EXCLUDED.name,
	    form_schema = EXCLUDED.form_schema,
	    prompt_preset = EXCLUDED.prompt_preset,
	    sample_image_url = EXCLUDED.sample_image_url,
	    is_active = true;`,
}

func Run(ctx context.Context, db *sql.DB) error {
	log.Println("[migration] running migrations...")
	for i, stmt := range migrations {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}
	log.Printf("[migration] %d migrations applied", len(migrations))

	log.Println("[migration] running seeds...")
	for i, stmt := range seedSQL {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("seed %d failed: %w", i+1, err)
		}
	}
	log.Printf("[migration] %d seeds applied", len(seedSQL))
	return nil
}
