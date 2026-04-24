package main

import (
	"log"

	"image-play/internal/config"
	http "image-play/internal/http"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	r := http.NewRouter(cfg.JWTSecret)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
