package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/scenes"
)

type ClientConfig struct {
	BrandSlogan string            `json:"brand_slogan"`
	Pricing     map[string]string `json:"pricing"`
	SceneOrder  []string          `json:"scene_order"`
}

func ClientConfigHandler(repo SceneTemplateLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		sceneOrder, err := runnableSceneOrder(c.Request.Context(), repo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load client config"})
			return
		}

		config := ClientConfig{
			BrandSlogan: "Play with your images",
			Pricing: map[string]string{
				"single": "1.00",
				"pack10": "8.00",
			},
			SceneOrder: sceneOrder,
		}
		c.JSON(http.StatusOK, config)
	}
}

func runnableSceneOrder(ctx context.Context, repo SceneTemplateLister) ([]string, error) {
	sceneOrder := make([]string, 0, len(scenes.SupportedSceneOrder()))
	for _, sceneKey := range scenes.SupportedSceneOrder() {
		items, err := repo.ListActiveByScene(ctx, sceneKey)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.IsRunnable() {
				sceneOrder = append(sceneOrder, sceneKey)
				break
			}
		}
	}
	return sceneOrder, nil
}
