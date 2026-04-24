package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/generation"
)

type CreateGenerationRequest struct {
	ClientRequestID string            `json:"client_request_id" binding:"required"`
	SceneKey        string            `json:"scene_key" binding:"required"`
	TemplateKey     string            `json:"template_key" binding:"required"`
	Fields          map[string]string `json:"fields"`
	SourceAssetID   *int64            `json:"source_asset_id,omitempty"`
}

type CreateGenerationResponse struct {
	GenerationID int64 `json:"generation_id"`
}

func CreateGenerationHandler(svc *generation.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		uid, ok := userID.(int64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		var req CreateGenerationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := svc.CreateGeneration(c.Request.Context(), generation.CreateGenerationInput{
			UserID:          uid,
			ClientRequestID: req.ClientRequestID,
			SceneKey:        req.SceneKey,
			TemplateKey:     req.TemplateKey,
			Fields:          req.Fields,
			SourceAssetID:   req.SourceAssetID,
		})
		if err != nil {
			if err == generation.ErrActiveGenerationExists {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "created"})
	}
}
