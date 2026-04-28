package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppEnv          string `yaml:"app_env"`
	Port            string `yaml:"port"`
	DatabaseURL     string `yaml:"database_url"`
	JWTSecret       string `yaml:"jwt_secret"`
	WechatAppID        string `yaml:"wechat_app_id"`
	WechatAppSecret    string `yaml:"wechat_app_secret"`
	OSSEndpoint        string `yaml:"oss_endpoint"`
	OSSBucket          string `yaml:"oss_bucket"`
	OSSAccessKeyID     string `yaml:"oss_access_key_id"`
	OSSAccessKeySecret string `yaml:"oss_access_key_secret"`
	DMXAPIKey          string `yaml:"dmx_api_key"`
	DMXAPIBaseURL      string `yaml:"dmx_api_base_url"`
	DMXModel           string `yaml:"dmx_model"`
}

func Load() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[config] warning: cannot read %s: %v, falling back to env vars\n", configPath, err)
		return loadFromEnv()
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "[config] warning: cannot parse %s: %v, falling back to env vars\n", configPath, err)
		return loadFromEnv()
	}

	// Allow env vars to override config file values
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.AppEnv = v
	}
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("WECHAT_APP_ID"); v != "" {
		cfg.WechatAppID = v
	}
	if v := os.Getenv("WECHAT_APP_SECRET"); v != "" {
		cfg.WechatAppSecret = v
	}
	if v := os.Getenv("OSS_ENDPOINT"); v != "" {
		cfg.OSSEndpoint = v
	}
	if v := os.Getenv("OSS_BUCKET"); v != "" {
		cfg.OSSBucket = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_ID"); v != "" {
		cfg.OSSAccessKeyID = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_SECRET"); v != "" {
		cfg.OSSAccessKeySecret = v
	}
	if v := os.Getenv("DMX_API_KEY"); v != "" {
		cfg.DMXAPIKey = v
	}
	if v := os.Getenv("DMX_API_BASE_URL"); v != "" {
		cfg.DMXAPIBaseURL = v
	}
	if v := os.Getenv("DMX_MODEL"); v != "" {
		cfg.DMXModel = v
	}

	return &cfg
}

func loadFromEnv() *Config {
	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		WechatAppID:        getEnv("WECHAT_APP_ID", ""),
		WechatAppSecret:    getEnv("WECHAT_APP_SECRET", ""),
		OSSEndpoint:        getEnv("OSS_ENDPOINT", ""),
		OSSBucket:          getEnv("OSS_BUCKET", ""),
		OSSAccessKeyID:     getEnv("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: getEnv("OSS_ACCESS_KEY_SECRET", ""),
		DMXAPIKey:          getEnv("DMX_API_KEY", ""),
		DMXAPIBaseURL:      getEnv("DMX_API_BASE_URL", "https://api.moonshot.cn/v1"),
		DMXModel:           getEnv("DMX_MODEL", "kimi-k2.6"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
