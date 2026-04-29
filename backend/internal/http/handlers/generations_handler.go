package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/generation"
)

type CreateGenerationRequest struct {
	ClientRequestID string            `json:"client_request_id" binding:"required"`
	SceneKey        string            `json:"scene_key"`
	TemplateKey     string            `json:"template_key"`
	Fields          map[string]string `json:"fields"`
	Prompt          string            `json:"prompt"`
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

		id, err := svc.CreateGeneration(c.Request.Context(), generation.CreateGenerationInput{
			UserID:          uid,
			ClientRequestID: req.ClientRequestID,
			SceneKey:        req.SceneKey,
			TemplateKey:     req.TemplateKey,
			Fields:          req.Fields,
			Prompt:          req.Prompt,
			SourceAssetID:   req.SourceAssetID,
		})
		if err != nil {
			if errors.Is(err, generation.ErrActiveGenerationExists) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, generation.ErrUnsupportedScene) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, generation.ErrTemplateNotAvailable) || errors.Is(err, generation.ErrTemplatePresetInvalid) {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
				return
			}
			log.Printf("CreateGeneration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"generation_id": id})
	}
}
