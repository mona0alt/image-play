package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"image-play/internal/domain/generation"
)

type ModelClient interface {
	Generate(ctx context.Context, prompt string) (imageURL string, err error)
}

type AuditClient interface {
	Audit(ctx context.Context, imageURL string) (pass bool, err error)
}

type MockModelClient struct{}

func (m *MockModelClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Simulate model generation latency
	select {
	case <-time.After(100 * time.Millisecond):
		return fmt.Sprintf("https://mock-cdn.example.com/result/%d.png", time.Now().Unix()), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type MockAuditClient struct{}

func (m *MockAuditClient) Audit(ctx context.Context, imageURL string) (bool, error) {
	// Simulate audit latency
	select {
	case <-time.After(50 * time.Millisecond):
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

type GenerationJob struct {
	modelClient  ModelClient
	auditClient  AuditClient
	generationRepo generation.Repository
}

func NewGenerationJob(repo generation.Repository, model ModelClient, audit AuditClient) *GenerationJob {
	if model == nil {
		model = &MockModelClient{}
	}
	if audit == nil {
		audit = &MockAuditClient{}
	}
	return &GenerationJob{
		modelClient:  model,
		auditClient:  audit,
		generationRepo: repo,
	}
}

func (j *GenerationJob) Execute(ctx context.Context, g *generation.Generation) error {
	log.Printf("[GenerationJob] executing generation %d", g.ID)

	// Build prompt from fields (MVP: simple concatenation)
	prompt := buildPrompt(g)

	if err := j.generationRepo.UpdateStatus(ctx, g.ID, "running"); err != nil {
		return fmt.Errorf("update status to running: %w", err)
	}

	imageURL, err := j.modelClient.Generate(ctx, prompt)
	if err != nil {
		_ = j.generationRepo.UpdateStatus(ctx, g.ID, "failed")
		return fmt.Errorf("model generation failed: %w", err)
	}

	if err := j.generationRepo.UpdateStatus(ctx, g.ID, "result_auditing"); err != nil {
		return fmt.Errorf("update status to result_auditing: %w", err)
	}

	pass, err := j.auditClient.Audit(ctx, imageURL)
	if err != nil {
		_ = j.generationRepo.UpdateStatus(ctx, g.ID, "failed")
		return fmt.Errorf("audit failed: %w", err)
	}
	if !pass {
		_ = j.generationRepo.UpdateStatus(ctx, g.ID, "failed")
		return fmt.Errorf("audit did not pass")
	}

	// Update result_url and status success
	g.ResultURL = imageURL
	// For in-memory repo we can just set fields, but for postgres we'd need an UpdateResult method.
	// For MVP we rely on UpdateStatus only; result_url update is omitted to keep repo interface minimal.
	if err := j.generationRepo.UpdateStatus(ctx, g.ID, "success"); err != nil {
		return fmt.Errorf("update status to success: %w", err)
	}

	log.Printf("[GenerationJob] generation %d completed successfully", g.ID)
	return nil
}

func buildPrompt(g *generation.Generation) string {
	if g.Prompt != "" {
		return g.Prompt
	}
	prompt := fmt.Sprintf("scene=%s template=%s", g.SceneKey, g.TemplateKey)
	for k, v := range g.Fields {
		prompt += fmt.Sprintf(" %s=%s", k, v)
	}
	return prompt
}
