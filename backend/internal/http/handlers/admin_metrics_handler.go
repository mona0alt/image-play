package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardMetrics struct {
	SceneClicks       map[string]int `json:"scene_clicks"`
	GenerationSuccess map[string]int `json:"generation_success"`
	Payments          map[string]int `json:"payments"`
}

func DashboardMetricsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := DashboardMetrics{
			SceneClicks:       make(map[string]int),
			GenerationSuccess: make(map[string]int),
			Payments:          make(map[string]int),
		}

		// Scene clicks: COUNT of tracking_events WHERE event = 'scene_clicked' GROUP BY scene_key
		rows, err := db.QueryContext(c.Request.Context(), `
			SELECT COALESCE(payload->>'scene_key', 'unknown'), COUNT(*)
			FROM tracking_events
			WHERE event = 'scene_clicked'
			GROUP BY payload->>'scene_key'
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query scene clicks"})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var sceneKey string
			var count int
			if err := rows.Scan(&sceneKey, &count); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan scene clicks"})
				return
			}
			metrics.SceneClicks[sceneKey] = count
		}

		// Generation success: COUNT of generations WHERE status = 'success' GROUP BY scene_key
		rows, err = db.QueryContext(c.Request.Context(), `
			SELECT scene_key, COUNT(*)
			FROM generations
			WHERE status = 'success'
			GROUP BY scene_key
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query generation success"})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var sceneKey string
			var count int
			if err := rows.Scan(&sceneKey, &count); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan generation success"})
				return
			}
			metrics.GenerationSuccess[sceneKey] = count
		}

		// Payments: COUNT of orders WHERE status = 'paid' GROUP BY DATE(created_at)
		rows, err = db.QueryContext(c.Request.Context(), `
			SELECT DATE(created_at), COUNT(*)
			FROM orders
			WHERE status = 'paid'
			GROUP BY DATE(created_at)
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query payments"})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var date string
			var count int
			if err := rows.Scan(&date, &count); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan payments"})
				return
			}
			metrics.Payments[date] = count
		}

		c.JSON(http.StatusOK, metrics)
	}
}
