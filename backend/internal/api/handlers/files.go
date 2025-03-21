package handlers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
)

func (h *Handler) processUploadedFile(ctx context.Context, file *multipart.FileHeader) (*models.File, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Calculate hash
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return nil, err
	}
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Reset file pointer
	if _, err := src.Seek(0, 0); err != nil {
		return nil, err
	}

	// Generate S3 key
	s3Key := path.Join("packages", fileHash, file.Filename)

	// Upload to S3
	if err := h.storage.Upload(ctx, s3Key, src); err != nil {
		return nil, err
	}

	// Generate download URL
	downloadURL, err := h.storage.GetSignedURL(ctx, s3Key, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &models.File{
		Name:        file.Filename,
		Size:        file.Size,
		Hash:        fileHash,
		ContentType: file.Header.Get("Content-Type"),
		DownloadURL: downloadURL,
		UploadedAt:  time.Now(),
	}, nil
}
