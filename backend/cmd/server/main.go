package main

import (
	"context"
	"log"

	"github.com/Parallels/pd-ai-agents-registry/internal/api/routes"
	"github.com/Parallels/pd-ai-agents-registry/internal/config"
	"github.com/Parallels/pd-ai-agents-registry/internal/database/seeder"
	"github.com/Parallels/pd-ai-agents-registry/internal/db/mongodb"
	"github.com/Parallels/pd-ai-agents-registry/internal/logger"
)

// @title           Parallels AI Registry API
// @version         1.0
// @description     API Server for Parallels AI Registry
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.parallels.com/support
// @contact.email  support@parallels.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @securityDefinitions.bearer BearerAuth
// @in header
// @name Authorization

func main() {
	// Initialize config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := logger.NewLogger(cfg.AppEnv)
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Initialize MongoDB client
	mongoClient, err := mongodb.NewClient(context.Background(), cfg.MongoDB)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	db := mongoClient.Database()
	seeder := seeder.NewDatabaseSeeder(db, cfg.Admin.Password)
	if err := seeder.Seed(context.Background()); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
	// Initialize router
	router := routes.NewRouter(cfg, logger, mongoClient)

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "80"
	}
	logger.Info("Starting server", "port", port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", err)
	}
}
