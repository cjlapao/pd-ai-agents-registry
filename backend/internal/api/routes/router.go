package routes

import (
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/api/handlers"
	"github.com/Parallels/pd-ai-agents-registry/internal/api/middleware"
	"github.com/Parallels/pd-ai-agents-registry/internal/config"
	"github.com/Parallels/pd-ai-agents-registry/internal/db/mongodb"
	"github.com/Parallels/pd-ai-agents-registry/internal/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg *config.Config, logger *logger.Logger, db *mongodb.Client) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	// Initialize handlers
	h, err := handlers.NewHandler(cfg, logger, db)
	if err != nil {
		logger.Fatal("Failed to initialize handlers", err)
	}
	auth := middleware.NewAuthMiddleware(cfg.JWT.Secret)

	// Configure rate limiter
	downloadRateLimit := middleware.RateLimit(middleware.RateLimitConfig{
		RequestsPerSecond: 1,               // 1 request per second
		BurstSize:         5,               // Allow bursts of up to 5 requests
		ExpiryTime:        time.Minute * 5, // Clean up visitors after 5 minutes
	})

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/register", h.Register)

		// Protected routes
		packages := v1.Group("/packages")
		packages.Use(auth.JWT())
		{
			packages.GET("", h.ListPackages)
			packages.GET("/:name", h.GetPackage)
			packages.GET("/:name/versions", h.ListVersions)
			packages.GET("/:name/versions/:version", h.GetVersion)

			// Protected with API key
			upload := packages.Group("")
			upload.Use(auth.APIKey())
			{
				upload.POST("/:name/versions/:version/upload", h.UploadPackage)
				upload.DELETE("/:name/versions/:version/:filename", h.DeletePackage)
			}
		}

		// Download route (public with rate limiting)
		v1.GET("/download/:name/:version/:filename", downloadRateLimit, h.DownloadPackage)

		// Update routes
		updates := v1.Group("/updates")
		{
			// Public routes
			updates.GET("", h.ListUpdates)
			updates.GET("/latest/:platform/:arch", h.GetLatestUpdate)
			updates.GET("/download/:version/:platform/:arch/:filename", downloadRateLimit, h.DownloadUpdate)
			updates.GET("/latest", h.GetLatestVersionInfo)

			// Protected routes (admin only)
			adminUpdates := updates.Group("")
			adminUpdates.Use(auth.JWT())
			{
				adminUpdates.POST("/:version/:platform/:arch", h.UploadUpdate)
				adminUpdates.DELETE("/:version/:platform/:arch", h.DeleteUpdate)
			}
		}
	}

	return router
}
