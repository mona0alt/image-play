package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/assets"
)

func UploadIntentHandler(assetService *assets.Service) gin.HandlerFunc {
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

		resp, err := assetService.CreateUploadIntent(c.Request.Context(), uid)
		if err != nil {
			log.Printf("CreateUploadIntent error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
