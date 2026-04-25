package http

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/assets"
	"image-play/internal/domain/billing"
	"image-play/internal/domain/generation"
	"image-play/internal/domain/tracking"
	"image-play/internal/domain/user"
	"image-play/internal/http/handlers"
	"image-play/internal/http/middleware"
	"image-play/internal/repository/postgres"
)

func NewRouter(db *sql.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	billingRepo := postgres.NewBillingRepo(db)
	billingSvc := billing.NewService(billingRepo)
	userRepo := postgres.NewUserRepo(db)
	userSvc := user.NewService(userRepo)
	templateRepo := postgres.NewSceneTemplateRepo(db)

	r.POST("/api/auth/login", handlers.LoginHandler(jwtSecret, userSvc))
	r.GET("/api/configs/client", handlers.ClientConfigHandler(templateRepo))
	r.GET("/api/scenes/:scene_key/templates", handlers.ListSceneTemplatesHandler(templateRepo))

	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(jwtSecret))
	authorized.GET("/me", handlers.MeHandler(userRepo))

	assetRepo := postgres.NewAssetRepo(db)
	assetSvc := assets.NewService(assetRepo)
	authorized.POST("/assets/upload-intent", handlers.UploadIntentHandler(assetSvc))

	genRepo := postgres.NewGenerationRepo(db)
	genSvc := generation.NewService(genRepo, templateRepo)
	authorized.POST("/generations", handlers.CreateGenerationHandler(genSvc))
	authorized.GET("/packages", handlers.PackagesHandler(billingSvc))
	authorized.POST("/orders", handlers.CreateOrderHandler(billingSvc))
	authorized.GET("/history", handlers.HistoryHandlerV2(genRepo))

	trackingRepo := postgres.NewTrackingRepo(db)
	trackingSvc := tracking.NewService(trackingRepo)
	authorized.POST("/tracking/events", handlers.TrackingEventsHandler(trackingSvc))

	r.POST("/api/payments/callback", handlers.PaymentCallbackHandler(billingSvc))

	// Admin routes — TODO: add admin role check middleware in production
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(jwtSecret))
	admin.GET("/metrics", handlers.DashboardMetricsHandler(db))
	admin.PUT("/templates/:id/toggle", handlers.ToggleTemplateHandler(db))

	return r
}
