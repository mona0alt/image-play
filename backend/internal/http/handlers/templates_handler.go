package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/scenes"
)

type SceneTemplateLister interface {
	ListActiveByScene(ctx context.Context, sceneKey string) ([]scenes.Template, error)
}

type sceneTemplateResponse struct {
	Key            string            `json:"key"`
	Name           string            `json:"name"`
	SceneKey       string            `json:"scene_key"`
	FormSchema     scenes.FormSchema `json:"form_schema"`
	SampleImageURL string            `json:"sample_image_url,omitempty"`
}

func ListSceneTemplatesHandler(repo SceneTemplateLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		sceneKey := c.Param("scene_key")
		if !scenes.IsSupportedScene(sceneKey) {
			c.JSON(http.StatusOK, gin.H{"items": []sceneTemplateResponse{}})
			return
		}

		items, err := repo.ListActiveByScene(c.Request.Context(), sceneKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
			return
		}

		resp := make([]sceneTemplateResponse, 0, len(items))
		for _, item := range items {
			if !item.IsRunnable() {
				continue
			}
			resp = append(resp, sceneTemplateResponse{
				Key:            item.TemplateKey,
				Name:           item.Name,
				SceneKey:       item.SceneKey,
				FormSchema:     item.FormSchema,
				SampleImageURL: item.SampleImageURL,
			})
		}

		c.JSON(http.StatusOK, gin.H{"items": resp})
	}
}
