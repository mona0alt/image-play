package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/tracking"
)

type TrackEventRequest struct {
	Event   string         `json:"event" binding:"required"`
	Payload map[string]any `json:"payload"`
}

type TrackingService interface {
	TrackEvent(ctx context.Context, userID int64, event string, payload map[string]any) error
}

func TrackingEventsHandler(svc *tracking.Service) gin.HandlerFunc {
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

		var req TrackEventRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := svc.TrackEvent(c.Request.Context(), uid, req.Event, req.Payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
