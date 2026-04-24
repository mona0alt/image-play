package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"image-play/internal/config"
	"image-play/internal/repository/postgres"
	"image-play/internal/worker"
)

func main() {
	log.Println("worker started")
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}
	repo := postgres.NewGenerationRepo(db)
	runner := worker.NewRunner(repo, nil)
	runner.Run(context.Background())
}
