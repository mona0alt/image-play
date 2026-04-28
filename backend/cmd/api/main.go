package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"image-play/internal/config"
	http "image-play/internal/http"
	"image-play/internal/infrastructure/storage"
	"image-play/internal/infrastructure/wechat"
	"image-play/internal/migration"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.WechatAppID == "" || cfg.WechatAppSecret == "" {
		log.Println("Warning: WECHAT_APP_ID or WECHAT_APP_SECRET is empty")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migration.Run(context.Background(), db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	var signer storage.Signer = storage.NoopSigner{}
	if cfg.OSSEndpoint != "" && cfg.OSSBucket != "" && cfg.OSSAccessKeyID != "" && cfg.OSSAccessKeySecret != "" {
		ossSigner, err := storage.NewOSSSigner(cfg.OSSEndpoint, cfg.OSSBucket, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
		if err != nil {
			log.Fatalf("oss signer init failed: %v", err)
		}
		signer = ossSigner
	} else {
		log.Println("Warning: OSS config missing, image URLs will not be signed")
	}

	wxClient := wechat.NewClient(cfg.WechatAppID, cfg.WechatAppSecret)
	r := http.NewRouter(db, cfg, wxClient, signer)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
