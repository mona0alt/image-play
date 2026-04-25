package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"image-play/internal/domain/user"
)

type LoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type User struct {
	ID        int64  `json:"id"`
	Openid    string `json:"openid"`
	Balance   int64  `json:"balance"`
	FreeQuota int64  `json:"free_quota"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

func LoginHandler(jwtSecret string, userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		account, _, err := userSvc.GetOrCreateByMockCode(c.Request.Context(), req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": strconv.FormatInt(account.ID, 10),
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		})

		accessToken, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		c.JSON(http.StatusOK, LoginResponse{
			AccessToken: accessToken,
			User: User{
				ID:        account.ID,
				Openid:    account.OpenID,
				Balance:   int64(account.Balance),
				FreeQuota: int64(account.FreeQuota),
			},
		})
	}
}

func MeHandler(userRepo user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("user_id")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		account, err := userRepo.GetByID(c.Request.Context(), uid)
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
			Openid:    account.OpenID,
			Balance:   int64(account.Balance),
			FreeQuota: int64(account.FreeQuota),
		})
	}
}
