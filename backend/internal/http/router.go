package http

import (
	"github.com/gin-gonic/gin"
	"image-play/internal/http/handlers"
	"image-play/internal/http/middleware"
)

func NewRouter(jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/api/auth/login", handlers.LoginHandler(jwtSecret))
	r.GET("/api/configs/client", handlers.ClientConfigHandler)

	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(jwtSecret))
	authorized.GET("/me", handlers.MeHandler)

	return r
}
