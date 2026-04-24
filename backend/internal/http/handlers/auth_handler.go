package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"image-play/internal/domain/billing"
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

func LoginHandler(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Mock WeChat login: any code returns a mock user
		user := User{
			ID:        1,
			Openid:    "mock-openid-" + req.Code,
			Balance:   0,
			FreeQuota: 3,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": strconv.FormatInt(user.ID, 10),
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
			User:        user,
		})
	}
}

func MeHandler(billingRepo billing.Repository) gin.HandlerFunc {
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

		user, err := billingRepo.GetUser(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}

		c.JSON(http.StatusOK, User{
			ID:        user.ID,
			Openid:    "",
			Balance:   int64(user.Balance),
			FreeQuota: int64(user.FreeQuota),
		})
	}
}

