package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
	"github.com/gin-gonic/gin"
)

// @Summary Download package file
// @Description Download a specific file from a package version
// @Tags packages
// @Accept json
// @Produce octet-stream
// @Param name path string true "Package name"
// @Param version path string true "Version"
// @Param filename path string true "Filename"
// @Success 200 {file} binary
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /download/{name}/{version}/{filename} [get]
func (h *Handler) DownloadPackage(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := c.Param("filename")

	// Get package
	pkg, err := h.db.GetPackage(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get package", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve package"})
		return
	}
	if pkg == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Package not found"})
		return
	}

	// Get version
	ver, err := h.db.GetVersion(c.Request.Context(), pkg.ID, version)
	if err != nil {
		h.logger.Error("Failed to get version", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve version"})
		return
	}
	if ver == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Version not found"})
		return
	}

	// Find file in version
	var fileInfo *models.File
	for _, f := range ver.Files {
		if f.Name == filename {
			fileInfo = &f
			break
		}
	}
	if fileInfo == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "File not found"})
		return
	}

	// Generate S3 key
	s3Key := fmt.Sprintf("packages/%s/%s/%s", name, version, filename)

	// Check if file exists in S3
	exists, err := h.storage.Exists(c.Request.Context(), s3Key)
	if err != nil {
		h.logger.Error("Failed to check file existence", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check file existence"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "File not found in storage"})
		return
	}

	// Get signed URL for download
	signedURL, err := h.storage.GetSignedURL(c.Request.Context(), s3Key, 15*time.Minute)
	if err != nil {
		h.logger.Error("Failed to generate signed URL", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate download URL"})
		return
	}

	// Redirect to signed URL
	c.Redirect(http.StatusTemporaryRedirect, signedURL)
}
