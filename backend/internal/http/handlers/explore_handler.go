package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExploreItem struct {
	ID           int64       `json:"id"`
	User         ExploreUser `json:"user"`
	ImageURL     string      `json:"image_url"`
	ThumbnailURL string      `json:"thumbnail_url"`
	Prompt       string      `json:"prompt"`
	SceneKey     string      `json:"scene_key"`
	LikeCount    int64       `json:"like_count"`
	IsLiked      bool        `json:"is_liked"`
	Description  string      `json:"description"`
	CreatedAt    string      `json:"created_at"`
}

type ExploreUser struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

func ExploreFeedHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.Query("page")
		pageSizeStr := c.Query("page_size")

		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize < 1 {
			pageSize = 10
		}

		query := `
			SELECT g.id, g.user_id, g.scene_key, g.template_key, g.result_url, g.prompt, g.created_at,
				u.nickname, u.avatar_url,
				COALESCE(l.cnt, 0) as like_count
			FROM generations g
			JOIN users u ON u.id = g.user_id
			LEFT JOIN (
				SELECT generation_id, COUNT(*) as cnt FROM likes GROUP BY generation_id
			) l ON l.generation_id = g.id
			WHERE g.status = 'success' AND g.result_url IS NOT NULL AND g.result_url != ''
			ORDER BY g.created_at DESC
			LIMIT $1 OFFSET $2
		`
		offset := (page - 1) * pageSize
		rows, err := db.QueryContext(c.Request.Context(), query, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
			return
		}
		defer rows.Close()

		items := []ExploreItem{}
		for rows.Next() {
			var item ExploreItem
			var userID int64
			var createdAt sql.NullTime
			var templateKey string
			err := rows.Scan(
				&item.ID, &userID, &item.SceneKey, &templateKey, &item.ImageURL, &item.Prompt, &createdAt,
				&item.User.Nickname, &item.User.AvatarURL,
				&item.LikeCount,
			)
			if err != nil {
				continue
			}
			item.User.ID = strconv.FormatInt(userID, 10)
			item.ThumbnailURL = item.ImageURL
			item.Description = item.Prompt
			if createdAt.Valid {
				item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z")
			}

			// Check if current user liked this item
			if uid, exists := c.Get("user_id"); exists {
				if userIDVal, ok := uid.(int64); ok {
					var existsLike bool
					_ = db.QueryRowContext(c.Request.Context(),
						"SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND generation_id = $2)",
						userIDVal, item.ID,
					).Scan(&existsLike)
					item.IsLiked = existsLike
				}
			}

			// Fallback nickname/avatar
			if item.User.Nickname == "" {
				item.User.Nickname = "用户" + item.User.ID
			}
			if item.User.AvatarURL == "" {
				item.User.AvatarURL = "https://api.dicebear.com/7.x/avataaars/svg?seed=" + item.User.ID
			}

			items = append(items, item)
		}

		var total int64
		countQuery := `SELECT COUNT(*) FROM generations WHERE status = 'success' AND result_url IS NOT NULL AND result_url != ''`
		_ = db.QueryRowContext(c.Request.Context(), countQuery).Scan(&total)

		c.JSON(http.StatusOK, gin.H{
			"items": items,
			"pagination": gin.H{
				"page":      page,
				"page_size": pageSize,
				"total":     total,
				"has_more":  int64(page*pageSize) < total,
			},
		})
	}
}

type LikeRequest struct {
	GenerationID int64  `json:"generation_id"`
	Action       string `json:"action"`
}

func ExploreLikeHandler(db *sql.DB) gin.HandlerFunc {
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

		var req LikeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.Action != "like" && req.Action != "unlike" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
			return
		}

		if req.Action == "like" {
			_, err := db.ExecContext(c.Request.Context(),
				"INSERT INTO likes (user_id, generation_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				uid, req.GenerationID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like"})
				return
			}
		} else {
			_, err := db.ExecContext(c.Request.Context(),
				"DELETE FROM likes WHERE user_id = $1 AND generation_id = $2",
				uid, req.GenerationID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike"})
				return
			}
		}

		var count int64
		_ = db.QueryRowContext(c.Request.Context(),
			"SELECT COUNT(*) FROM likes WHERE generation_id = $1", req.GenerationID,
		).Scan(&count)

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"like_count": count,
		})
	}
}
