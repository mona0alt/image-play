package generation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateGenerationRejectsWhenActiveJobExists(t *testing.T) {
	repo := NewInMemoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Create first active generation
	err1 := svc.CreateGeneration(ctx, CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "office-pro",
		Fields:          map[string]string{},
	})
	require.NoError(t, err1)

	// Try to create second active generation for same user
	err2 := svc.CreateGeneration(ctx, CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-2",
		SceneKey:        "invitation",
		TemplateKey:     "wedding-classic",
		Fields:          map[string]string{},
	})
	require.ErrorIs(t, err2, ErrActiveGenerationExists)
}
