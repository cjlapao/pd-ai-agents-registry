package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UploadUpdate handles uploading a new application update
// @Summary Upload application update
// @Description Upload a new version of the application
// @Tags updates
// @Accept multipart/form-data
// @Produce json
// @Param version path string true "Version number"
// @Param platform path string true "Platform (windows, darwin, linux)"
// @Param arch path string true "Architecture (x86_64, i686, armv7, aarch64)"
// @Param file formData file true "Update file"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/updates/{version}/{platform}/{arch} [post]
// @Security ApiKeyAuth
func (h *Handler) UploadUpdate(c *gin.Context) {
	version := c.Param("version")
	platform := c.Param("platform")
	arch := c.Param("arch")
	if arch == "x86" {
		arch = "i686"
	}
	if arch == "arm64" {
		arch = "aarch64"
	}

	// Validate platform
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid platform. Must be one of: windows, darwin, linux"})
		return
	}

	// Validate arch
	if !isValidArch(arch) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid arch. Must be one of: x86, x64, arm64"})
		return
	}

	version = strings.TrimPrefix(version, "v")

	// Check if version already exists for this platform and architecture
	collection := h.db.Database().Collection("updates")
	count, err := collection.CountDocuments(
		c.Request.Context(),
		bson.M{
			"version":  version,
			"platform": platform,
			"arch":     arch,
		},
	)
	if err != nil {
		h.logger.Error("Failed to check for existing update", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check for existing update"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, ErrorResponse{Error: fmt.Sprintf("Update for version %s on %s/%s already exists", version, platform, arch)})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get file from form", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to get file from form"})
		return
	}
	defer file.Close()

	// get signature file from form
	signature, _, err := c.Request.FormFile("signature")
	if err != nil {
		h.logger.Error("Failed to get signature file from form", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to get signature file from form"})
		return
	}
	defer signature.Close()

	// Read signature file content
	signatureContent, err := io.ReadAll(signature)
	if err != nil {
		h.logger.Error("Failed to read signature file", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to read signature file"})
		return
	}

	// Reset reader position for later upload
	_, _ = signature.Seek(0, 0)

	// upload update file to s3
	updateKey := fmt.Sprintf("updates/%s/%s/%s/%s", version, platform, arch, header.Filename)
	if err := h.storage.Upload(c.Request.Context(), updateKey, file); err != nil {
		h.logger.Error("Failed to upload update file to S3", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to upload update file to S3"})
		return
	}

	// Create download URL
	downloadURL := fmt.Sprintf("/api/v1/updates/download/%s/%s/%s/%s", version, platform, arch, header.Filename)

	// Create update record in database
	now := time.Now()
	update := models.Update{
		Version:     version,
		Platform:    platform,
		Arch:        arch,
		Filename:    header.Filename,
		FileSize:    header.Size,
		Signature:   string(signatureContent),
		ReleaseDate: now,
		DownloadURL: downloadURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Get notes from form if provided
	if notes := c.PostForm("notes"); notes != "" {
		update.Notes = notes
	}

	// Insert into database
	_, err = collection.InsertOne(c.Request.Context(), update)
	if err != nil {
		h.logger.Error("Failed to save update record", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save update record"})
		return
	}

	// Update the LatestVersion document
	if err := h.updateLatestVersionDocument(c.Request.Context(), update); err != nil {
		h.logger.Error("Failed to update latest version document", err)
		// Don't return an error to the client, as the upload was successful
	}

	// Return success
	c.JSON(http.StatusCreated, gin.H{
		"message":      "Update uploaded successfully",
		"version":      version,
		"platform":     platform,
		"filename":     header.Filename,
		"download_url": downloadURL,
	})
}

// GetLatestUpdate returns information about the latest update
// @Summary Get latest update
// @Description Get information about the latest update for a platform
// @Tags updates
// @Produce json
// @Param platform path string true "Platform (windows, macos, linux)"
// @Success 200 {object} models.UpdateMetadata
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/updates/latest/{platform}/{arch} [get]
func (h *Handler) GetLatestUpdate(c *gin.Context) {
	platform := c.Param("platform")
	arch := c.Param("arch")

	// Validate platform
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid platform. Must be one of: windows, macos, linux"})
		return
	}

	// Validate arch
	if !isValidArch(arch) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid arch. Must be one of: x86, x64, arm64"})
		return
	}

	// Find the latest update for the platform
	collection := h.db.Database().Collection("updates")

	// Sort by version in descending order (assuming semantic versioning)
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

	var update models.Update
	err := collection.FindOne(
		c.Request.Context(),
		bson.M{"platform": platform, "arch": arch},
		opts,
	).Decode(&update)
	if err != nil {
		h.logger.Error("Failed to find latest update", err)
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No updates found for this platform"})
		return
	}

	// Return update metadata
	metadata := models.UpdateMetadata{
		Version:     update.Version,
		Platform:    update.Platform,
		Arch:        update.Arch,
		ReleaseDate: update.ReleaseDate,
		Notes:       update.Notes,
		DownloadURL: update.DownloadURL,
	}

	c.JSON(http.StatusOK, metadata)
}

// DownloadUpdate downloads an update file
// @Summary Download update
// @Description Download an update file
// @Tags updates
// @Produce octet-stream
// @Param version path string true "Version number"
// @Param platform path string true "Platform (windows, macos, linux)"
// @Param arch path string true "Architecture (x86, x64, arm64)"
// @Param filename path string true "Filename"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/updates/download/{version}/{platform}/{arch}/{filename} [get]
func (h *Handler) DownloadUpdate(c *gin.Context) {
	version := c.Param("version")
	platform := c.Param("platform")
	arch := c.Param("arch")
	filename := c.Param("filename")

	// Validate platform
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid platform. Must be one of: windows, darwin, linux"})
		return
	}

	// Validate arch
	if !isValidArch(arch) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid arch. Must be one of: x86_64, i686, armv7, aarch64"})
		return
	}

	// Trim v prefix if present
	version = strings.TrimPrefix(version, "v")

	// Generate S3 key
	key := fmt.Sprintf("updates/%s/%s/%s/%s", version, platform, arch, filename)

	// Check if file exists
	exists, err := h.storage.Exists(c.Request.Context(), key)
	if err != nil {
		h.logger.Error("Failed to check if update file exists", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check if update file exists"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Update file not found"})
		return
	}

	// Get file size
	size, err := h.storage.Size(c.Request.Context(), key)
	if err != nil {
		h.logger.Error("Failed to get file size", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get file size"})
		return
	}

	c.Header("Content-Length", strconv.Itoa(int(size)))

	// Get file from S3
	reader, err := h.storage.Download(c.Request.Context(), key)
	if err != nil {
		h.logger.Error("Failed to download update file", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to download update file"})
		return
	}
	defer reader.Close()

	// Set content disposition header
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Set content type based on file extension
	contentType := getContentType(filename)
	c.Header("Content-Type", contentType)

	// Stream file to response
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		h.logger.Error("Failed to stream update file", err)
	}
}

// ListUpdates lists all available updates
// @Summary List updates
// @Description List all available updates
// @Tags updates
// @Produce json
// @Success 200 {array} models.UpdateMetadata
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/updates [get]
func (h *Handler) ListUpdates(c *gin.Context) {
	collection := h.db.Database().Collection("updates")

	// Find all updates, sorted by version and platform
	cursor, err := collection.Find(
		c.Request.Context(),
		bson.M{},
		options.Find().SetSort(bson.D{
			{Key: "platform", Value: 1},
			{Key: "version", Value: -1},
		}),
	)
	if err != nil {
		h.logger.Error("Failed to list updates", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list updates"})
		return
	}
	defer cursor.Close(c.Request.Context())

	// Decode updates
	var updates []models.Update
	if err := cursor.All(c.Request.Context(), &updates); err != nil {
		h.logger.Error("Failed to decode updates", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode updates"})
		return
	}

	// Convert to metadata
	var metadata []models.UpdateMetadata
	for _, update := range updates {
		metadata = append(metadata, models.UpdateMetadata{
			Version:     update.Version,
			Platform:    update.Platform,
			Arch:        update.Arch,
			ReleaseDate: update.ReleaseDate,
			Notes:       update.Notes,
			DownloadURL: update.DownloadURL,
		})
	}

	c.JSON(http.StatusOK, metadata)
}

// DeleteUpdate deletes an update
// @Summary Delete update
// @Description Delete an update and its files
// @Tags updates
// @Produce json
// @Param version path string true "Version number"
// @Param platform path string true "Platform (windows, darwin, linux)"
// @Param arch path string true "Architecture (x86, x64, arm64)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/updates/{version}/{platform}/{arch} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteUpdate(c *gin.Context) {
	version := c.Param("version")
	platform := c.Param("platform")
	arch := c.Param("arch")

	// Validate platform
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid platform. Must be one of: windows, darwin, linux"})
		return
	}

	// Validate arch
	if !isValidArch(arch) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid arch. Must be one of: x86, x64, arm64"})
		return
	}

	// Trim v prefix if present
	version = strings.TrimPrefix(version, "v")

	// Find the update to get filenames
	collection := h.db.Database().Collection("updates")
	var update models.Update
	err := collection.FindOne(
		c.Request.Context(),
		bson.M{
			"version":  version,
			"platform": platform,
			"arch":     arch,
		},
	).Decode(&update)
	if err != nil {
		h.logger.Error("Failed to find update", err)
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Update not found"})
		return
	}

	// Delete the update file from S3
	updateKey := fmt.Sprintf("updates/%s/%s/%s/%s", version, platform, arch, update.Filename)
	if err := h.storage.Delete(c.Request.Context(), updateKey); err != nil {
		h.logger.Error("Failed to delete update file from S3", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete update file"})
		return
	}

	// Delete the signature file from S3 if it exists
	if update.Signature != "" {
		signatureKey := fmt.Sprintf("updates/%s/%s/%s/%s", version, platform, arch, update.Signature)
		if err := h.storage.Delete(c.Request.Context(), signatureKey); err != nil {
			h.logger.Error("Failed to delete signature file from S3", err)
			// Continue anyway, as this is not critical
		}
	}

	// Delete the update from the database
	result, err := collection.DeleteOne(
		c.Request.Context(),
		bson.M{
			"version":  version,
			"platform": platform,
			"arch":     arch,
		},
	)
	if err != nil {
		h.logger.Error("Failed to delete update from database", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete update from database"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Update not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Update deleted successfully",
		"version":  version,
		"platform": platform,
		"arch":     arch,
	})
}

// Helper functions
func isValidPlatform(platform string) bool {
	validPlatforms := map[string]bool{
		"windows": true,
		"darwin":  true,
		"linux":   true,
	}
	return validPlatforms[platform]
}

func isValidArch(arch string) bool {
	if arch == "x86" {
		arch = "i686"
	}
	if arch == "arm64" {
		arch = "aarch64"
	}
	validArchs := map[string]bool{
		"x86_64":  true,
		"i686":    true,
		"armv7":   true,
		"aarch64": true,
	}
	return validArchs[arch]
}

func generateAllPlatformKeys() []string {
	platforms := []string{"windows", "darwin", "linux"}
	archs := []string{"x86_64", "i686", "armv7", "aarch64"}
	keys := []string{}
	for _, platform := range platforms {
		for _, arch := range archs {
			keys = append(keys, fmt.Sprintf("%s-%s", platform, arch))
		}
	}
	return keys
}

func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".zip":
		return "application/zip"
	case ".exe":
		return "application/octet-stream"
	case ".dmg":
		return "application/x-apple-diskimage"
	case ".deb":
		return "application/vnd.debian.binary-package"
	case ".rpm":
		return "application/x-rpm"
	case ".tar.gz":
		return "application/x-gzip"
	case ".tar.xz":
		return "application/x-xz"
	case ".tar.bz2":
		return "application/x-bzip2"
	default:
		return "application/octet-stream"
	}
}

// updateLatestVersionDocument updates the LatestVersion document with the latest update information
func (h *Handler) updateLatestVersionDocument(ctx context.Context, update models.Update) error {
	collection := h.db.Database().Collection("latest_version")

	// Get the current latest version document
	var latestVersion models.LatestVersion
	err := collection.FindOne(ctx, bson.M{}).Decode(&latestVersion)
	if err != nil && err != mongo.ErrNoDocuments {
		return fmt.Errorf("error finding latest version document: %w", err)
	}

	// If no document exists or the new version is newer, update the version and notes
	if err == mongo.ErrNoDocuments || compareVersions(update.Version, latestVersion.Version) > 0 {
		latestVersion.Version = update.Version
		latestVersion.Notes = update.Notes
		latestVersion.PubDate = update.ReleaseDate.Format(time.RFC3339)
	} else if update.Version != latestVersion.Version {
		// If this is an older version, don't update the document
		return nil
	}

	// Update the platform-specific information
	platformKey := fmt.Sprintf("%s-%s", update.Platform, update.Arch)
	platformInfo := models.LatestVersionPlatform{
		URL: update.DownloadURL,
	}

	// Add signature URL if available
	if update.Signature != "" {
		platformInfo.Signature = update.Signature
	}

	// Ensure all platform keys are present
	for _, platformKey := range generateAllPlatformKeys() {
		if _, ok := latestVersion.Platforms[platformKey]; !ok {
			if latestVersion.Platforms == nil {
				latestVersion.Platforms = make(map[string]models.LatestVersionPlatform)
			}
			latestVersion.Platforms[platformKey] = models.LatestVersionPlatform{}
		}
	}

	if update.Notes != "" {
		latestVersion.Notes = update.Notes
	}

	// Update the appropriate platform field based on the platform key
	latestVersion.Platforms[platformKey] = platformInfo

	// Upsert the document
	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(
		ctx,
		bson.M{}, // Empty filter to match any document
		bson.M{"$set": latestVersion},
		opts,
	)
	if err != nil {
		return fmt.Errorf("error upserting latest version document: %w", err)
	}

	return nil
}

// compareVersions compares two semantic version strings
// Returns:  1 if v1 > v2
//
//	-1 if v1 < v2
//	 0 if v1 == v2
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split versions into components
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	// Compare each component
	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		v1Num, _ := strconv.Atoi(v1Parts[i])
		v2Num, _ := strconv.Atoi(v2Parts[i])

		if v1Num > v2Num {
			return 1
		} else if v1Num < v2Num {
			return -1
		}
	}

	// If we get here, the common parts are equal, so the longer one is greater
	if len(v1Parts) > len(v2Parts) {
		return 1
	} else if len(v1Parts) < len(v2Parts) {
		return -1
	}

	return 0
}

// GetLatestVersionInfo returns the latest version information for all platforms
// @Summary Get latest version info
// @Description Get information about the latest version for all platforms
// @Tags updates
// @Produce json
// @Success 200 {object} models.LatestVersion
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/updates/latest [get]
func (h *Handler) GetLatestVersionInfo(c *gin.Context) {
	collection := h.db.Database().Collection("latest_version")

	var latestVersion models.LatestVersion
	err := collection.FindOne(c.Request.Context(), bson.M{}).Decode(&latestVersion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "No updates available"})
			return
		}
		h.logger.Error("Failed to get latest version info", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get latest version info"})
		return
	}

	// Sort the platforms by platform key
	platforms := make(map[string]models.LatestVersionPlatform)
	for _, platform := range generateAllPlatformKeys() {
		if info, ok := latestVersion.Platforms[platform]; ok {
			info.URL = fmt.Sprintf("%s%s", h.cfg.GetBaseURL(), info.URL)
			platforms[platform] = info
		}
	}

	latestVersion.Platforms = platforms
	c.JSON(http.StatusOK, latestVersion)
}
