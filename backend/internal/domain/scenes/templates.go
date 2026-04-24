package scenes

type TemplatePreset struct {
	BasePrompt string
	StyleWords []string
}

var SceneTemplates = map[string]map[string]TemplatePreset{
	SceneInvitation: {
		"wedding-classic": {
			BasePrompt: "婚礼请柬，优雅浪漫风格，白色与金色主色调，花卉装饰边框，新人姓名与婚礼信息清晰呈现，高品质印刷质感。",
			StyleWords: []string{"elegant", "romantic", "floral", "classic"},
		},
		"wedding-modern": {
			BasePrompt: "现代简约婚礼请柬，极简排版，莫兰迪色系，几何线条装饰，高级感。",
			StyleWords: []string{"minimalist", "modern", "geometric", "muted tones"},
		},
	},
	ScenePortrait: {
		"office-pro": {
			BasePrompt: "职业形象照，商务正装，纯色背景，专业灯光，自信表情，高清人像摄影。",
			StyleWords: []string{"professional", "business", "portrait", "studio lighting"},
		},
		"creative-artist": {
			BasePrompt: "艺术家肖像，创意光影，个性表达，艺术氛围，独特构图。",
			StyleWords: []string{"artistic", "creative", "expressive", "dramatic lighting"},
		},
	},
	SceneFestival: {
		"spring-festival": {
			BasePrompt: "春节贺卡，喜庆红色主题，金色祥云与烟花，传统中国元素，福字与灯笼，温馨团圆氛围。",
			StyleWords: []string{"festive", "traditional", "Chinese New Year", "red and gold"},
		},
	},
	SceneTshirt: {
		"streetwear": {
			BasePrompt: "街头潮流T恤图案，涂鸦风格，大胆配色，城市文化元素，年轻活力。",
			StyleWords: []string{"streetwear", "graffiti", "bold", "urban"},
		},
	},
	ScenePoster: {
		"concert": {
			BasePrompt: "演唱会海报，强烈视觉冲击，霓虹灯光效果，音乐人形象，动感排版。",
			StyleWords: []string{"concert", "neon", "dynamic", "music"},
		},
	},
}
