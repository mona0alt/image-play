-- Seed data for scene_templates

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'scene_templates') THEN
    INSERT INTO scene_templates (scene_key, template_key, name, form_schema, prompt_preset, sample_image_url, is_active)
    VALUES
      ('portrait', 'office-pro', '通勤职业', '[{"name":"subject_name","label":"拍摄对象","type":"text","required":true}]', '{"base_prompt":"职业形象照，商务正装，纯色背景，专业灯光，自信表情，高清人像摄影。","style_words":["professional","business","portrait","studio lighting"]}', 'https://example.com/portrait-office-pro.png', TRUE),
      ('festival', 'spring-festival', '春节祝福', '[{"name":"title","label":"标题","type":"text","required":true}]', '{"base_prompt":"春节贺卡，喜庆红色主题，金色祥云与烟花，传统中国元素，福字与灯笼，温馨团圆氛围。","style_words":["festive","traditional","Chinese New Year","red and gold"]}', 'https://example.com/festival-spring-festival.png', TRUE),
      ('invitation', 'wedding-classic', '婚礼请柬', '[{"name":"host_name","label":"主办人","type":"text","required":true}]', '{"base_prompt":"婚礼请柬，优雅浪漫风格，白色与金色主色调，花卉装饰边框，新人姓名与婚礼信息清晰呈现，高品质印刷质感。","style_words":["elegant","romantic","floral","classic"]}', 'https://example.com/invitation-wedding-classic.png', TRUE),
      ('tshirt', 'streetwear', '街头潮流', '[{"name":"theme","label":"主题","type":"text","required":true}]', '{"base_prompt":"街头潮流T恤图案，涂鸦风格，大胆配色，城市文化元素，年轻活力。","style_words":["streetwear","graffiti","bold","urban"]}', 'https://example.com/tshirt-streetwear.png', TRUE),
      ('poster', 'concert', '演唱会海报', '[{"name":"title","label":"标题","type":"text","required":true}]', '{"base_prompt":"演唱会海报，强烈视觉冲击，霓虹灯光效果，音乐人形象，动感排版。","style_words":["concert","neon","dynamic","music"]}', 'https://example.com/poster-concert.png', TRUE)
    ON CONFLICT (scene_key, template_key) DO UPDATE
    SET name = EXCLUDED.name,
        form_schema = EXCLUDED.form_schema,
        prompt_preset = EXCLUDED.prompt_preset,
        sample_image_url = EXCLUDED.sample_image_url;
  END IF;
END $$;
