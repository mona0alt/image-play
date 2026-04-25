package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ToggleTemplateRequest struct {
	Active bool `json:"active"`
}

func ToggleTemplateHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
			return
		}

		var req ToggleTemplateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		res, err := db.ExecContext(c.Request.Context(),
			"UPDATE scene_templates SET is_active = $1 WHERE id = $2",
			req.Active, id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template"})
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check update result"})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id, "active": req.Active})
	}
}
