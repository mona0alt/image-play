package config

import "os"

type Config struct {
	AppEnv           string
	Port             string
	DatabaseURL      string
	JWTSecret        string
	WechatAppID      string
	WechatAppSecret  string
}

func Load() *Config {
	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		WechatAppID:     getEnv("WECHAT_APP_ID", ""),
		WechatAppSecret: getEnv("WECHAT_APP_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
