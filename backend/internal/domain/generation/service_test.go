package generation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"image-play/internal/domain/scenes"
)

func TestCreateGenerationRejectsWhenActiveJobExists(t *testing.T) {
	repo := newInMemoryRepo()
	templates := newInMemoryTemplateRepo()
	templates.Set(&scenes.Template{
		SceneKey:    "portrait",
		TemplateKey: "office-pro",
		PromptPreset: scenes.PromptPreset{
			BasePrompt: "职业形象照",
		},
		Active: true,
	})
	templates.Set(&scenes.Template{
		SceneKey:    "invitation",
		TemplateKey: "wedding-classic",
		PromptPreset: scenes.PromptPreset{
			BasePrompt: "婚礼请柬",
		},
		Active: true,
	})
	svc := NewService(repo, templates)
	ctx := context.Background()

	// Create first active generation
	_, err1 := svc.CreateGeneration(ctx, CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "office-pro",
		Fields:          map[string]string{},
	})
	require.NoError(t, err1)

	// Try to create second active generation for same user
	_, err2 := svc.CreateGeneration(ctx, CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-2",
		SceneKey:        "invitation",
		TemplateKey:     "wedding-classic",
		Fields:          map[string]string{},
	})
	require.ErrorIs(t, err2, ErrActiveGenerationExists)
}

func TestCreateGenerationRejectsInactiveTemplate(t *testing.T) {
	repo := newInMemoryRepo()
	templates := newInMemoryTemplateRepo()
	svc := NewService(repo, templates)

	_, err := svc.CreateGeneration(context.Background(), CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "office-pro",
		Fields:          map[string]string{},
	})

	require.ErrorIs(t, err, ErrTemplateNotAvailable)
	require.Empty(t, repo.generations)
}

func TestCreateGenerationRejectsUnsupportedScene(t *testing.T) {
	repo := newInMemoryRepo()
	templates := newInMemoryTemplateRepo()
	templates.Set(&scenes.Template{
		SceneKey:    "unknown-scene",
		TemplateKey: "some-template",
		PromptPreset: scenes.PromptPreset{
			BasePrompt: "should never be used",
		},
		Active: true,
	})
	svc := NewService(repo, templates)

	_, err := svc.CreateGeneration(context.Background(), CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-1",
		SceneKey:        "unknown-scene",
		TemplateKey:     "some-template",
		Fields:          map[string]string{},
	})

	require.ErrorIs(t, err, ErrUnsupportedScene)
	require.Empty(t, repo.generations)
}

func TestCreateGenerationRejectsTemplateWithInvalidPromptPreset(t *testing.T) {
	repo := newInMemoryRepo()
	templates := newInMemoryTemplateRepo()
	templates.Set(&scenes.Template{
		SceneKey:     "portrait",
		TemplateKey:  "office-pro",
		PromptPreset: scenes.PromptPreset{},
		Active:       true,
	})
	svc := NewService(repo, templates)

	_, err := svc.CreateGeneration(context.Background(), CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "office-pro",
		Fields:          map[string]string{},
	})

	require.ErrorIs(t, err, ErrTemplatePresetInvalid)
	require.Empty(t, repo.generations)
}
