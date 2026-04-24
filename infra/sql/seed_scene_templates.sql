-- Seed data for scene_templates

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'scene_templates') THEN
    INSERT INTO scene_templates (scene_key, template_key, name, form_schema, prompt_preset)
    VALUES ('portrait', 'office-pro', '通勤职业', '{}', '{}')
    ON CONFLICT (scene_key, template_key) DO NOTHING;
  END IF;
END $$;
