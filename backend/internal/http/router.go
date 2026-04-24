package http

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/assets"
	"image-play/internal/domain/generation"
	"image-play/internal/http/handlers"
	"image-play/internal/http/middleware"
	"image-play/internal/repository/postgres"
)

func NewRouter(db *sql.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/api/auth/login", handlers.LoginHandler(jwtSecret))
	r.GET("/api/configs/client", handlers.ClientConfigHandler)

	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(jwtSecret))
	authorized.GET("/me", handlers.MeHandler)

	assetRepo := postgres.NewAssetRepo(db)
	assetSvc := assets.NewService(assetRepo)
	authorized.POST("/assets/upload-intent", handlers.UploadIntentHandler(assetSvc))

	genRepo := postgres.NewGenerationRepo(db)
	genSvc := generation.NewService(genRepo)
	authorized.POST("/generations", handlers.CreateGenerationHandler(genSvc))

	return r
}
