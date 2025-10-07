.PHONY: build install test clean dev help

# Binary name
BINARY_NAME=envswitch

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Version information from git
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags
LDFLAGS=-X github.com/hugofrely/envswitch/internal/version.Version=$(VERSION) \
        -X github.com/hugofrely/envswitch/internal/version.GitCommit=$(GIT_COMMIT) \
        -X github.com/hugofrely/envswitch/internal/version.BuildDate=$(BUILD_DATE)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) -v

install: build ## Install the binary to /usr/local/bin
	@echo "Installing to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed successfully!"

dev: ## Build and run in development mode
	@echo "Building and running..."
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) -v
	@./$(BUILD_DIR)/$(BINARY_NAME)

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "❌ golangci-lint not installed." && echo "Install with: brew install golangci-lint" && echo "Or: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin" && exit 1)
	@golangci-lint run --config .golangci.yml

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@which golangci-lint > /dev/null || (echo "❌ golangci-lint not installed." && echo "Install with: brew install golangci-lint" && exit 1)
	@golangci-lint run --fix --config .golangci.yml

fmt: ## Format code with gofmt
	@echo "Formatting code..."
	@gofmt -w .
	@$(GOCMD) mod tidy

check: ## Quick check (fmt, vet, build, test)
	@echo "Running quick checks..."
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) build
	@$(MAKE) test
	@echo "✅ Quick checks passed!"

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

install-lint: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@which brew > /dev/null && brew install golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.64.8
	@echo "✓ golangci-lint installed successfully!"

ci: fmt vet lint test-race ## Run all CI checks locally (format, vet, lint, test)
	@echo "✅ All CI checks passed!"

setup-hooks: ## Install git pre-commit hooks
	@echo "Installing git hooks..."
	@bash scripts/setup-hooks.sh

run: build ## Build and run
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Cross-compilation targets
build-linux: ## Build for Linux
	@echo "Building for Linux $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v

build-darwin: ## Build for macOS
	@echo "Building for macOS $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -v

build-windows: ## Build for Windows
	@echo "Building for Windows $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v

build-all: build-linux build-darwin build-windows ## Build for all platforms
	@echo "Built for all platforms!"

.DEFAULT_GOAL := help
