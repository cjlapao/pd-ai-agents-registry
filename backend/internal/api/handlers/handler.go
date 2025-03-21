package handlers

import (
	"github.com/Parallels/pd-ai-agents-registry/internal/config"
	"github.com/Parallels/pd-ai-agents-registry/internal/db/mongodb"
	"github.com/Parallels/pd-ai-agents-registry/internal/logger"
	"github.com/Parallels/pd-ai-agents-registry/internal/services/storage"
)

type Handler struct {
	cfg     *config.Config
	logger  *logger.Logger
	db      *mongodb.Client
	storage *storage.S3Service
}

func NewHandler(cfg *config.Config, logger *logger.Logger, db *mongodb.Client) (*Handler, error) {
	// Initialize S3 storage service
	s3Service, err := storage.NewS3Service(cfg.S3)
	if err != nil {
		return nil, err
	}

	return &Handler{
		cfg:     cfg,
		logger:  logger,
		db:      db,
		storage: s3Service,
	}, nil
}
