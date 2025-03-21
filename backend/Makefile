# Variables
APP_NAME := pd-ai-agent-registry
DOCKER_REGISTRY := cjlapao
IMAGE_NAME := $(DOCKER_REGISTRY)/$(APP_NAME)
IMAGE_TAG := latest

# Go related variables
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOLINT := golangci-lint

# Main build directory
BUILD_DIR := build

# Default target
.PHONY: all
all: clean deps lint test build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "Clean complete"

# Install project dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@$(GOGET) -v ./...
	@$(GOMOD) tidy
	@echo "Dependencies installed"

# Verify dependencies
.PHONY: verify
verify:
	@echo "Verifying dependencies..."
	@$(GOMOD) verify
	@echo "Dependencies verified"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...
	@echo "Tests complete"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not found, installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.55.2; \
		$(GOBIN)/$(GOLINT) run ./...; \
	fi
	@echo "Linting complete"

# Build the application
.PHONY: build
build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Build for specific platforms
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	@echo "Linux build complete: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

.PHONY: build-mac
build-mac:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	@echo "macOS build complete: $(BUILD_DIR)/$(APP_NAME)-darwin-amd64"

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	@echo "Windows build complete: $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe"

# Docker related targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image: $(IMAGE_NAME):$(IMAGE_TAG)..."
	@echo "Building for Linux AMD64..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) --platform linux/amd64 .
	@echo "Building for Linux ARM64..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) --platform linux/arm64 .
	@echo "Docker build complete"

.PHONY: docker-push
docker-push: docker-build
	@echo "Pushing Docker image: $(IMAGE_NAME):$(IMAGE_TAG)..."
	@echo "Pushing for Linux AMD64..."
	@docker push $(IMAGE_NAME):$(IMAGE_TAG) --platform linux/amd64
	@echo "Pushing for Linux ARM64..."
	@docker push $(IMAGE_NAME):$(IMAGE_TAG) --platform linux/arm64
	@echo "Docker push complete"

# Build and tag with specified version
.PHONY: docker-build-version
docker-build-version:
	@echo "Building Docker image with version tag: $(IMAGE_NAME):$(VERSION)..."
	@docker build -t $(IMAGE_NAME):$(VERSION) .
	@echo "Docker build with version tag complete"

.PHONY: docker-push-version
docker-push-version: docker-build-version
	@echo "Pushing Docker image with version tag: $(IMAGE_NAME):$(VERSION)..."
	@docker push $(IMAGE_NAME):$(VERSION)
	@echo "Docker push with version tag complete"

# Helm related targets
.PHONY: helm-lint
helm-lint:
	@echo "Linting Helm chart..."
	@helm lint ./helm/backend
	@echo "Helm lint complete"

.PHONY: helm-template
helm-template:
	@echo "Generating Helm template..."
	@helm template release-name ./helm/backend > helm-template.yaml
	@echo "Helm template generated: helm-template.yaml"

# Run the application
.PHONY: run
run: build
	@echo "Running application..."
	@./$(BUILD_DIR)/$(APP_NAME)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Clean, install dependencies, lint, test, and build"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  verify           - Verify dependencies"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Run linter"
	@echo "  build            - Build application"
	@echo "  build-linux      - Build for Linux"
	@echo "  build-mac        - Build for macOS"
	@echo "  build-windows    - Build for Windows"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-push      - Build and push Docker image"
	@echo "  docker-build-version - Build Docker image with specific version (use VERSION=x.y.z)"
	@echo "  docker-push-version  - Push Docker image with specific version (use VERSION=x.y.z)"
	@echo "  helm-lint        - Lint Helm chart"
	@echo "  helm-template    - Generate Helm template"
	@echo "  run              - Run the application"
	@echo "  help             - Show this help message"

# Default target (if no arguments provided)
.DEFAULT_GOAL := help 