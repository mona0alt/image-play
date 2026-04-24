package worker

import (
	"context"
	"log"
	"time"

	"image-play/internal/domain/generation"
	"image-play/internal/worker/jobs"
)

type Runner struct {
	generationRepo generation.Repository
	job            *jobs.GenerationJob
	tickInterval   time.Duration
}

func NewRunner(repo generation.Repository, job *jobs.GenerationJob) *Runner {
	if job == nil {
		job = jobs.NewGenerationJob(repo, nil, nil)
	}
	return &Runner{
		generationRepo: repo,
		job:            job,
		tickInterval:   time.Second,
	}
}

func (r *Runner) Run(ctx context.Context) {
	log.Println("[Worker] runner started")
	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] shutting down")
			return
		default:
		}

		job, err := r.generationRepo.Dequeue(ctx)
		if err != nil {
			log.Printf("[Worker] dequeue error: %v", err)
			time.Sleep(r.tickInterval)
			continue
		}
		if job == nil {
			time.Sleep(r.tickInterval)
			continue
		}

		log.Printf("[Worker] processing job %d", job.ID)
		if err := r.job.Execute(ctx, job); err != nil {
			log.Printf("[Worker] job %d failed: %v", job.ID, err)
		}
	}
}
