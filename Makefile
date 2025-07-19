# ZetMem MCP Server Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop setup dev deps lint fmt vet

# Variables
BINARY_NAME=zetmem-server
DOCKER_IMAGE=zetmem/mcp-server
VERSION?=1.0.0
CONFIG_FILE?=config/development.yaml

# Default target
help: ## Show this help message
	@echo "ZetMem MCP Server - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
setup: ## Initial project setup
	@echo "Setting up ZetMem MCP Server development environment..."
	@cp .env.example .env
	@echo "✅ Created .env file (please update with your API keys)"
	@go mod tidy
	@echo "✅ Downloaded Go dependencies"
	@mkdir -p data
	@echo "✅ Created data directory"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit .env file with your API keys"
	@echo "2. Run 'make dev' to start development environment"

deps: ## Download and tidy Go dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

build: deps ## Build the server binary
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 go build -ldflags="-X main.version=$(VERSION)" -o $(BINARY_NAME) cmd/server/main.go
	@echo "✅ Built $(BINARY_NAME)"

run: build ## Run the server locally
	@echo "Starting ZetMem MCP Server..."
	@./$(BINARY_NAME) -config $(CONFIG_FILE)

dev: ## Start development environment with Docker Compose
	@echo "Starting development environment..."
	@docker-compose up -d chromadb redis
	@echo "✅ ChromaDB and Redis started"
	@echo "Starting server..."
	@go run cmd/server/main.go -config config/development.yaml

# Testing
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(VERSION)..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "✅ Built Docker image"

docker-run: ## Run with Docker Compose
	@echo "Starting ZetMem MCP Server with Docker Compose..."
	@docker-compose up -d
	@echo "✅ Services started"
	@echo "Server: http://localhost:8080"
	@echo "ChromaDB: http://localhost:8000"
	@echo "Metrics: http://localhost:9090"

docker-stop: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	@docker-compose down
	@echo "✅ Services stopped"

docker-logs: ## View Docker logs
	@docker-compose logs -f zetmem-server

# Utilities
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@go clean
	@echo "✅ Cleaned"

reset-data: ## Reset all data (ChromaDB, Redis)
	@echo "Resetting all data..."
	@docker-compose down -v
	@rm -rf data/*
	@echo "✅ Data reset"

logs: ## View server logs (when running with docker-compose)
	@docker-compose logs -f zetmem-server

status: ## Check service status
	@echo "Service Status:"
	@docker-compose ps

# Installation
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "✅ Installed"

# Release
release: clean test docker-build ## Build release artifacts
	@echo "Building release $(VERSION)..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/$(BINARY_NAME)-linux-amd64 cmd/server/main.go
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/$(BINARY_NAME)-darwin-amd64 cmd/server/main.go
	@GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/server/main.go
	@echo "✅ Release artifacts built in dist/"

# Phase 2 commands
test-phase2: build ## Run Phase 2 feature tests
	@echo "Running Phase 2 tests..."
	@python3 scripts/test_phase2.py

metrics: ## View Prometheus metrics
	@echo "Opening metrics endpoint..."
	@curl -s http://localhost:9090/metrics | head -20

health: ## Check service health
	@echo "Checking service health..."
	@curl -s http://localhost:9090/health

evolution: ## Trigger manual memory evolution
	@echo "Triggering memory evolution..."
	@python3 -c "import json; print(json.dumps({'jsonrpc': '2.0', 'id': 1, 'method': 'tools/call', 'params': {'name': 'evolve_memory_network', 'arguments': {'trigger_type': 'manual', 'scope': 'recent'}}}))" | ./zetmem-server -config config/development.yaml

# Quick commands
up: docker-run ## Alias for docker-run
down: docker-stop ## Alias for docker-stop
restart: docker-stop docker-run ## Restart services
