package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"image-play/internal/infrastructure/storage"
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

func ExploreFeedHandler(db *sql.DB, signer storage.Signer) gin.HandlerFunc {
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
			SELECT ea.id, ea.image_url, ea.scene_key, ea.prompt, ea.created_at,
				COALESCE(l.cnt, 0) as like_count
			FROM explore_assets ea
			LEFT JOIN (
				SELECT explore_asset_id, COUNT(*) as cnt FROM explore_likes GROUP BY explore_asset_id
			) l ON l.explore_asset_id = ea.id
			ORDER BY ea.created_at DESC
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
			var createdAt sql.NullTime
			err := rows.Scan(
				&item.ID, &item.ImageURL, &item.SceneKey, &item.Prompt, &createdAt,
				&item.LikeCount,
			)
			if err != nil {
				continue
			}
			if signer != nil {
				item.ImageURL = signer.SignImageURL(item.ImageURL)
			}
			item.ThumbnailURL = item.ImageURL
			item.Description = ""
			if createdAt.Valid {
				item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z")
			}

			// Check if current user liked this item
			if uid, exists := c.Get("user_id"); exists {
				if userIDVal, ok := uid.(int64); ok {
					var existsLike bool
					_ = db.QueryRowContext(c.Request.Context(),
						"SELECT EXISTS(SELECT 1 FROM explore_likes WHERE user_id = $1 AND explore_asset_id = $2)",
						userIDVal, item.ID,
					).Scan(&existsLike)
					item.IsLiked = existsLike
				}
			}

			item.User = ExploreUser{
				ID:        "0",
				Nickname:  "",
				AvatarURL: "",
			}

			items = append(items, item)
		}

		var total int64
		countQuery := `SELECT COUNT(*) FROM explore_assets`
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
	ExploreAssetID int64  `json:"explore_asset_id"`
	Action         string `json:"action"`
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
				"INSERT INTO explore_likes (user_id, explore_asset_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				uid, req.ExploreAssetID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like"})
				return
			}
		} else {
			_, err := db.ExecContext(c.Request.Context(),
				"DELETE FROM explore_likes WHERE user_id = $1 AND explore_asset_id = $2",
				uid, req.ExploreAssetID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike"})
				return
			}
		}

		var count int64
		_ = db.QueryRowContext(c.Request.Context(),
			"SELECT COUNT(*) FROM explore_likes WHERE explore_asset_id = $1", req.ExploreAssetID,
		).Scan(&count)

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"like_count": count,
		})
	}
}
