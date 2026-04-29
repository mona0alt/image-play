package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"image-play/internal/domain/generation"
	"image-play/internal/domain/scenes"
)

type stubGenerationRepo struct {
	generation *generation.Generation
}

func (r *stubGenerationRepo) Create(_ context.Context, g *generation.Generation) error {
	r.generation = g
	return nil
}

func (r *stubGenerationRepo) GetActiveByUser(_ context.Context, userID int64) (*generation.Generation, error) {
	return nil, nil
}

func (r *stubGenerationRepo) Dequeue(_ context.Context) (*generation.Generation, error) {
	return nil, nil
}

func (r *stubGenerationRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	if r.generation == nil || r.generation.ID != id {
		return errors.New("generation not found")
	}
	r.generation.Status = status
	r.generation.UpdatedAt = time.Now()
	return nil
}

func (r *stubGenerationRepo) UpdateResult(_ context.Context, id int64, status, resultURL string) error {
	if r.generation == nil || r.generation.ID != id {
		return errors.New("generation not found")
	}
	r.generation.Status = status
	r.generation.ResultURL = resultURL
	r.generation.UpdatedAt = time.Now()
	return nil
}

func (r *stubGenerationRepo) ListByUser(_ context.Context, userID int64) ([]*generation.Generation, error) {
	return nil, nil
}

func (r *stubGenerationRepo) ListSuccess(_ context.Context, _, _ int) ([]*generation.Generation, int64, error) {
	return nil, 0, nil
}

type stubTemplateLoader struct {
	template *scenes.Template
	err      error
}

func (l *stubTemplateLoader) GetActiveTemplate(_ context.Context, sceneKey, templateKey string) (*scenes.Template, error) {
	if l.err != nil {
		return nil, l.err
	}
	if l.template != nil && l.template.SceneKey == sceneKey && l.template.TemplateKey == templateKey {
		return l.template, nil
	}
	return nil, nil
}

type captureModelClient struct {
	prompt string
	calls  int
}

func (m *captureModelClient) Generate(_ context.Context, prompt string) (string, error) {
	m.calls++
	m.prompt = prompt
	return "https://example.com/result.png", nil
}

type passAuditClient struct{}

func (a *passAuditClient) Audit(_ context.Context, imageURL string) (bool, error) {
	return true, nil
}

func TestExecuteBuildsPromptFromTemplatePreset(t *testing.T) {
	repo := &stubGenerationRepo{
		generation: &generation.Generation{
			ID:          1,
			UserID:      2,
			SceneKey:    "portrait",
			TemplateKey: "office-pro",
			Fields: map[string]string{
				"subject_name": "张三",
			},
			Status: "running",
		},
	}
	templateLoader := &stubTemplateLoader{
		template: &scenes.Template{
			SceneKey:    "portrait",
			TemplateKey: "office-pro",
			PromptPreset: scenes.PromptPreset{
				BasePrompt: "职业形象照 preset",
				StyleWords: []string{"professional"},
			},
			Active: true,
		},
	}
	model := &captureModelClient{}
	job := NewGenerationJob(repo, templateLoader, model, &passAuditClient{}, nil)

	err := job.Execute(context.Background(), repo.generation)

	require.NoError(t, err)
	require.Contains(t, model.prompt, "职业形象照 preset")
	require.Contains(t, model.prompt, "professional")
	require.Contains(t, model.prompt, "张三")
	require.NotContains(t, model.prompt, "scene=portrait")
	require.Equal(t, "success", repo.generation.Status)
	require.Equal(t, "https://example.com/result.png", repo.generation.ResultURL)
}

func TestNewGenerationJobRequiresTemplateLookup(t *testing.T) {
	require.PanicsWithValue(t, "template lookup is required", func() {
		NewGenerationJob(&stubGenerationRepo{}, nil, &captureModelClient{}, nil, nil)
	})
}

func TestNewGenerationJobRequiresImageClient(t *testing.T) {
	require.PanicsWithValue(t, "image client is required", func() {
		NewGenerationJob(&stubGenerationRepo{}, &stubTemplateLoader{}, nil, nil, nil)
	})
}

func TestExecuteRejectsTemplateWithInvalidPromptPreset(t *testing.T) {
	repo := &stubGenerationRepo{
		generation: &generation.Generation{
			ID:          1,
			UserID:      2,
			SceneKey:    "portrait",
			TemplateKey: "office-pro",
			Fields: map[string]string{
				"subject_name": "张三",
			},
			Status: "running",
		},
	}
	templateLoader := &stubTemplateLoader{
		template: &scenes.Template{
			SceneKey:     "portrait",
			TemplateKey:  "office-pro",
			PromptPreset: scenes.PromptPreset{},
			Active:       true,
		},
	}
	model := &captureModelClient{}
	job := NewGenerationJob(repo, templateLoader, model, &passAuditClient{}, nil)

	err := job.Execute(context.Background(), repo.generation)

	require.EqualError(t, err, "template preset invalid")
	require.Equal(t, 0, model.calls)
	require.Equal(t, "failed", repo.generation.Status)
	require.Empty(t, repo.generation.ResultURL)
}
