package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/user"
)

func UpdateMeHandler(userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("user_id")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Nickname string `json:"nickname" binding:"required,min=2,max=20"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := userSvc.UpdateNickname(c.Request.Context(), uid, req.Nickname); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		account, err := userSvc.GetByID(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}
		if account == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, User{
			ID:        account.ID,
			Nickname:  account.Nickname,
			Balance:   int64(account.Balance),
			FreeQuota: int64(account.FreeQuota),
		})
	}
}
