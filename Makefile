.PHONY: all build clean test lint install help build-all

# Binary name
BINARY_NAME=certfix
VERSION?=1.0.0
BUILD_DIR=bin
DIST_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# Platforms
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

all: clean lint test build

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(DIST_DIR)/$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then \
				output_name=$$output_name.exe; \
			fi; \
			echo "Building $$platform/$$arch..."; \
			GOOS=$$platform GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $$output_name main.go; \
			if [ $$? -ne 0 ]; then \
				echo "Build failed for $$platform/$$arch"; \
			fi; \
		done; \
	done
	@echo "Cross-compilation complete. Binaries in $(DIST_DIR)/"

install: ## Install the binary to GOBIN
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) .
	@echo "Installation complete"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	$(GOCLEAN)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go fmt and go vet..."; \
		$(GOFMT) ./...; \
		$(GOCMD) vet ./...; \
	fi

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

run: build ## Build and run the application
	@$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode
	@$(GOCMD) run main.go

# Platform-specific builds
build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 main.go

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 main.go

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go

version: ## Display version
	@echo "Version: $(VERSION)"
