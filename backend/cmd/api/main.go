package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"image-play/internal/config"
	http "image-play/internal/http"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := http.NewRouter(db, cfg.JWTSecret)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
