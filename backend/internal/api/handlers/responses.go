package handlers

import "github.com/Parallels/pd-ai-agents-registry/internal/models"

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type UploadResponse struct {
	Message string       `json:"message"`
	File    *models.File `json:"file"`
}
