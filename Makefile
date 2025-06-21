# Makefile for Clean Route Backend Microservice

# Variables
BINARY_NAME=clean-route-backend
BUILD_DIR=build
DOCKER_IMAGE=clean-route-backend
DOCKER_TAG=latest
GO_FILES=$(shell find . -name "*.go" -type f)
MAIN_FILE=main.go

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_UNIX=$(BINARY_NAME)_unix

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Clean Route Backend Microservice - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development targets
.PHONY: run
run: ## Run the application locally
	@echo "🚀 Starting Clean Route Backend..."
	$(GOCMD) run $(MAIN_FILE)

.PHONY: run-watch
run-watch: ## Run the application with file watching (requires air)
	@echo "👀 Starting with file watching..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

.PHONY: dev
dev: ## Setup development environment
	@echo "🔧 Setting up development environment..."
	@if [ ! -f .envrc ]; then \
		echo "📝 Creating .envrc from example..."; \
		cp .envrc.example .envrc 2>/dev/null || echo "⚠️  .envrc.example not found, create .envrc manually"; \
	fi
	@echo "✅ Development environment ready!"
	@echo "📋 Next steps:"
	@echo "   1. Edit .envrc with your API keys"
	@echo "   2. Run 'direnv allow' to load environment"
	@echo "   3. Run 'make run' to start the server"

# Build targets
.PHONY: build
build: ## Build the application
	@echo "🔨 Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "🐧 Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_FILE)
	@echo "✅ Linux build complete: $(BUILD_DIR)/$(BINARY_UNIX)"

.PHONY: build-mac
build-mac: ## Build for macOS
	@echo "🍎 Building for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_mac $(MAIN_FILE)
	@echo "✅ macOS build complete: $(BUILD_DIR)/$(BINARY_NAME)_mac"

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "🪟 Building for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_FILE)
	@echo "✅ Windows build complete: $(BUILD_DIR)/$(BINARY_NAME).exe"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "🌍 Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	$(MAKE) build-linux
	$(MAKE) build-mac
	$(MAKE) build-windows
	@echo "✅ All platform builds complete!"

# Test targets
.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -v -cover ./...

.PHONY: test-coverage-html
test-coverage-html: ## Run tests with HTML coverage report
	@echo "📊 Generating HTML coverage report..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

.PHONY: test-benchmark
test-benchmark: ## Run benchmark tests
	@echo "⚡ Running benchmarks..."
	$(GOTEST) -bench=. ./...

# Code quality targets
.PHONY: fmt
fmt: ## Format Go code
	@echo "🎨 Formatting Go code..."
	$(GOCMD) fmt ./...

.PHONY: vet
vet: ## Vet Go code
	@echo "🔍 Vetting Go code..."
	$(GOCMD) vet ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "🔍 Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

.PHONY: check
check: ## Run all code quality checks
	@echo "✅ Running code quality checks..."
	$(MAKE) fmt
	$(MAKE) vet
	$(MAKE) lint

# Dependency targets
.PHONY: deps
deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "🔄 Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

.PHONY: deps-clean
deps-clean: ## Clean dependencies
	@echo "🧹 Cleaning dependencies..."
	$(GOCLEAN) -modcache

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --env-file .envrc $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	@echo "🛑 Stopping Docker container..."
	docker stop $$(docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG)) 2>/dev/null || echo "No containers running"

.PHONY: docker-clean
docker-clean: ## Clean Docker images
	@echo "🧹 Cleaning Docker images..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || echo "Image not found"

# Cleanup targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete!"

.PHONY: clean-all
clean-all: ## Clean everything including Docker
	@echo "🧹 Deep cleaning..."
	$(MAKE) clean
	$(MAKE) docker-clean
	$(MAKE) deps-clean
	@echo "✅ Deep cleanup complete!"

# Health check targets
.PHONY: health
health: ## Check service health
	@echo "💚 Checking service health..."
	@curl -f http://localhost:8080/health || echo "❌ Service not responding"

.PHONY: health-wait
health-wait: ## Wait for service to be healthy
	@echo "⏳ Waiting for service to be healthy..."
	@until curl -f http://localhost:8080/health > /dev/null 2>&1; do \
		echo "Waiting for service..."; \
		sleep 2; \
	done
	@echo "✅ Service is healthy!"

# Development workflow targets
.PHONY: setup
setup: ## Complete development setup
	@echo "🚀 Setting up development environment..."
	$(MAKE) deps
	$(MAKE) dev
	$(MAKE) check
	@echo "✅ Development environment ready!"

.PHONY: dev-workflow
dev-workflow: ## Development workflow (setup, test, run)
	@echo "🔄 Running development workflow..."
	$(MAKE) setup
	$(MAKE) test
	$(MAKE) run

.PHONY: ci
ci: ## CI/CD pipeline steps
	@echo "🔧 Running CI/CD pipeline..."
	$(MAKE) deps
	$(MAKE) check
	$(MAKE) test-coverage
	$(MAKE) build
	@echo "✅ CI/CD pipeline complete!"

# Documentation targets
.PHONY: docs
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@if command -v godoc > /dev/null; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "❌ godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
		exit 1; \
	fi

# Release targets
.PHONY: release
release: ## Create a release build
	@echo "🎉 Creating release build..."
	$(MAKE) clean
	$(MAKE) test
	$(MAKE) build-all
	@echo "✅ Release build complete!"

.PHONY: release-docker
release-docker: ## Create a release Docker image
	@echo "🎉 Creating release Docker image..."
	$(MAKE) test
	$(MAKE) docker-build
	@echo "✅ Release Docker image complete!"

# Utility targets
.PHONY: version
version: ## Show version information
	@echo "📋 Version Information:"
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build time: $$(date)"

.PHONY: info
info: ## Show project information
	@echo "📋 Project Information:"
	@echo "Binary name: $(BINARY_NAME)"
	@echo "Main file: $(MAIN_FILE)"
	@echo "Go files: $$(find . -name "*.go" | wc -l)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "🛠️ Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "✅ Development tools installed!" 