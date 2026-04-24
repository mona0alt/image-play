package jobs

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"image-play/internal/domain/billing"
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
	modelClient    ModelClient
	auditClient    AuditClient
	generationRepo generation.Repository
	billingSvc     *billing.Service
}

func NewGenerationJob(repo generation.Repository, model ModelClient, audit AuditClient, billingSvc *billing.Service) *GenerationJob {
	if model == nil {
		model = &MockModelClient{}
	}
	if audit == nil {
		audit = &MockAuditClient{}
	}
	return &GenerationJob{
		modelClient:    model,
		auditClient:    audit,
		generationRepo: repo,
		billingSvc:     billingSvc,
	}
}

func (j *GenerationJob) Execute(ctx context.Context, g *generation.Generation) error {
	log.Printf("[GenerationJob] executing generation %d", g.ID)

	// Build prompt from fields (MVP: simple concatenation)
	prompt := buildPrompt(g)

	imageURL, err := j.modelClient.Generate(ctx, prompt)
	if err != nil {
		_ = j.generationRepo.UpdateResult(ctx, g.ID, "failed", "")
		return fmt.Errorf("model generation failed: %w", err)
	}

	if err := j.generationRepo.UpdateStatus(ctx, g.ID, "result_auditing"); err != nil {
		return fmt.Errorf("update status to result_auditing: %w", err)
	}

	pass, err := j.auditClient.Audit(ctx, imageURL)
	if err != nil {
		_ = j.generationRepo.UpdateResult(ctx, g.ID, "failed", "")
		return fmt.Errorf("audit failed: %w", err)
	}
	if !pass {
		_ = j.generationRepo.UpdateResult(ctx, g.ID, "failed", "")
		return fmt.Errorf("audit did not pass")
	}

	if err := j.generationRepo.UpdateResult(ctx, g.ID, "success", imageURL); err != nil {
		return fmt.Errorf("update status to success: %w", err)
	}

	if j.billingSvc != nil {
		if err := j.billingSvc.ChargeGeneration(ctx, g.UserID, g.ID); err != nil {
			log.Printf("[GenerationJob] billing charge failed for generation %d: %v", g.ID, err)
		}
	}

	log.Printf("[GenerationJob] generation %d completed successfully", g.ID)
	return nil
}

func buildPrompt(g *generation.Generation) string {
	if g.Prompt != "" {
		return g.Prompt
	}
	prompt := fmt.Sprintf("scene=%s template=%s", g.SceneKey, g.TemplateKey)
	keys := make([]string, 0, len(g.Fields))
	for k := range g.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		prompt += fmt.Sprintf(" %s=%s", k, g.Fields[k])
	}
	return prompt
}
