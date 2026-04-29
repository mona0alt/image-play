package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"image-play/internal/config"
	"image-play/internal/domain/billing"
	"image-play/internal/infrastructure/llm"
	"image-play/internal/migration"
	"image-play/internal/repository/postgres"
	"image-play/internal/worker"
	"image-play/internal/worker/jobs"
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
	if err := migration.Run(context.Background(), db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	repo := postgres.NewGenerationRepo(db)
	templateRepo := postgres.NewSceneTemplateRepo(db)
	billingRepo := postgres.NewBillingRepo(db)
	billingSvc := billing.NewService(billingRepo)

	imageClient, err := llm.NewImageClient(llm.ImageConfig{
		APIKey:  cfg.LLM.Image.APIKey,
		BaseURL: cfg.LLM.Image.BaseURL,
		Model:   cfg.LLM.Image.Model,
		Timeout: cfg.LLM.Image.Timeout,
	})
	if err != nil {
		log.Fatalf("image client init failed: %v", err)
	}

	job := jobs.NewGenerationJob(repo, templateRepo, imageClient, nil, billingSvc)
	runner := worker.NewRunner(repo, job)
	runner.Run(context.Background())
}
