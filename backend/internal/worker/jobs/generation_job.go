package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"image-play/internal/domain/billing"
	"image-play/internal/domain/generation"
	"image-play/internal/domain/scenes"
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
	templateRepo   generation.TemplateLookup
	billingSvc     *billing.Service
}

func NewGenerationJob(repo generation.Repository, templateRepo generation.TemplateLookup, model ModelClient, audit AuditClient, billingSvc *billing.Service) *GenerationJob {
	if templateRepo == nil {
		panic("template lookup is required")
	}

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
		templateRepo:   templateRepo,
		billingSvc:     billingSvc,
	}
}

func (j *GenerationJob) Execute(ctx context.Context, g *generation.Generation) error {
	log.Printf("[GenerationJob] executing generation %d", g.ID)

	prompt, err := j.buildPrompt(ctx, g)
	if err != nil {
		_ = j.generationRepo.UpdateResult(ctx, g.ID, "failed", "")
		return err
	}

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

func (j *GenerationJob) buildPrompt(ctx context.Context, g *generation.Generation) (string, error) {
	template, err := j.templateRepo.GetActiveTemplate(ctx, g.SceneKey, g.TemplateKey)
	if err != nil {
		return "", fmt.Errorf("load active template: %w", err)
	}
	if template == nil {
		return "", fmt.Errorf("template not available")
	}

	return scenes.BuildPrompt(scenes.BuildInput{
		SceneKey:    g.SceneKey,
		TemplateKey: g.TemplateKey,
		Preset:      template.PromptPreset,
		Fields:      g.Fields,
	}), nil
}
