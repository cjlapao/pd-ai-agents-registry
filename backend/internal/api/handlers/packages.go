package handlers

import (
	"net/http"

	"github.com/Parallels/pd-ai-agents-registry/internal/models"
	"github.com/gin-gonic/gin"
)

// @Summary List all packages
// @Description Get a list of all available packages
// @Tags packages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Package
// @Failure 500 {object} ErrorResponse
// @Router /packages [get]
func (h *Handler) ListPackages(c *gin.Context) {
	packages, err := h.db.ListPackages(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list packages", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve packages"})
		return
	}
	c.JSON(http.StatusOK, packages)
}

// @Summary Get package details
// @Description Get details for a specific package
// @Tags packages
// @Accept json
// @Produce json
// @Param name path string true "Package name"
// @Security BearerAuth
// @Success 200 {object} models.Package
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /packages/{name} [get]
func (h *Handler) GetPackage(c *gin.Context) {
	name := c.Param("name")
	pkg, err := h.db.GetPackage(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get package", "error", err, "name", name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve package"})
		return
	}
	if pkg == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Package not found"})
		return
	}
	c.JSON(http.StatusOK, pkg)
}

// @Summary List package versions
// @Description Get all versions for a specific package
// @Tags packages
// @Accept json
// @Produce json
// @Param name path string true "Package name"
// @Security BearerAuth
// @Success 200 {array} models.Version
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /packages/{name}/versions [get]
func (h *Handler) ListVersions(c *gin.Context) {
	name := c.Param("name")

	// First get the package to get its ID
	pkg, err := h.db.GetPackage(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get package", "error", err, "name", name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve package"})
		return
	}
	if pkg == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Package not found"})
		return
	}

	versions, err := h.db.ListVersions(c.Request.Context(), pkg.ID)
	if err != nil {
		h.logger.Error("Failed to list versions", "error", err, "package", name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve versions"})
		return
	}
	c.JSON(http.StatusOK, versions)
}

// @Summary Get version details
// @Description Get details for a specific package version
// @Tags packages
// @Accept json
// @Produce json
// @Param name path string true "Package name"
// @Param version path string true "Version"
// @Security BearerAuth
// @Success 200 {object} models.Version
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /packages/{name}/versions/{version} [get]
func (h *Handler) GetVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	pkg, err := h.db.GetPackage(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get package", "error", err, "name", name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve package"})
		return
	}
	if pkg == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Package not found"})
		return
	}

	ver, err := h.db.GetVersion(c.Request.Context(), pkg.ID, version)
	if err != nil {
		h.logger.Error("Failed to get version", "error", err, "package", name, "version", version)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve version"})
		return
	}
	if ver == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Version not found"})
		return
	}
	c.JSON(http.StatusOK, ver)
}

// @Summary Upload package version
// @Description Upload a new package version
// @Tags packages
// @Accept multipart/form-data
// @Produce json
// @Param name path string true "Package name"
// @Param version path string true "Version"
// @Param file formData file true "Package file"
// @Security ApiKeyAuth
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /packages/{name}/versions/{version}/upload [post]
func (h *Handler) UploadPackage(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	// Get package
	pkg, err := h.db.GetPackage(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get package", "error", err, "name", name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve package"})
		return
	}
	if pkg == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Package not found"})
		return
	}

	// Get the file
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get uploaded file", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to get uploaded file"})
		return
	}

	// Process and store the file
	fileModel, err := h.processUploadedFile(c.Request.Context(), file)
	if err != nil {
		h.logger.Error("Failed to process file", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process uploaded file"})
		return
	}

	// Create or update version
	ver, err := h.db.GetVersion(c.Request.Context(), pkg.ID, version)
	if err != nil {
		h.logger.Error("Failed to get version", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check version"})
		return
	}

	if ver == nil {
		// Create new version
		ver = &models.Version{
			PackageID: pkg.ID,
			Version:   version,
			Files:     []models.File{*fileModel},
		}
		if err := h.db.CreateVersion(c.Request.Context(), ver); err != nil {
			h.logger.Error("Failed to create version", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create version"})
			return
		}
	} else {
		// Add file to existing version
		if err := h.db.AddFileToVersion(c.Request.Context(), pkg.ID, version, *fileModel); err != nil {
			h.logger.Error("Failed to add file to version", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add file to version"})
			return
		}
	}

	c.JSON(http.StatusOK, UploadResponse{
		Message: "File uploaded successfully",
		File:    fileModel,
	})
}

// @Summary Delete package file
// @Description Delete a file from a package version
// @Tags packages
// @Accept json
// @Produce json
// @Param name path string true "Package name"
// @Param version path string true "Version"
// @Param filename path string true "Filename"
// @Security ApiKeyAuth
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /packages/{name}/versions/{version}/{filename} [delete]
func (h *Handler) DeletePackage(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := c.Param("filename")

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

	if err := h.db.RemoveFileFromVersion(c.Request.Context(), pkg.ID, version, filename); err != nil {
		h.logger.Error("Failed to remove file", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove file"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "File deleted successfully"})
}
