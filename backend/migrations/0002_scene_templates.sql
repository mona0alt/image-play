-- Migration 0002_scene_templates: create scene_templates table

CREATE TABLE IF NOT EXISTS scene_templates (
    id BIGSERIAL PRIMARY KEY,
    scene_key VARCHAR(32) NOT NULL,
    template_key VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    form_schema JSONB NOT NULL,
    prompt_preset JSONB NOT NULL,
    sample_image_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE (scene_key, template_key)
);
