package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/generation"
)

type HistoryItem struct {
	ID          int64             `json:"id"`
	SceneKey    string            `json:"scene_key"`
	TemplateKey string            `json:"template_key"`
	Fields      map[string]string `json:"fields"`
	Status      string            `json:"status"`
	ResultURL   string            `json:"result_url"`
	CreatedAt   string            `json:"created_at"`
}

type GenerationLister interface {
	ListByUser(ctx context.Context, userID int64) ([]*generation.Generation, error)
}

func HistoryHandlerV2(lister GenerationLister) gin.HandlerFunc {
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

		gens, err := lister.ListByUser(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
			return
		}

		items := make([]HistoryItem, 0, len(gens))
		for _, g := range gens {
			items = append(items, HistoryItem{
				ID:          g.ID,
				SceneKey:    g.SceneKey,
				TemplateKey: g.TemplateKey,
				Fields:      g.Fields,
				Status:      g.Status,
				ResultURL:   g.ResultURL,
				CreatedAt:   strconv.FormatInt(g.CreatedAt.Unix(), 10),
			})
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}
