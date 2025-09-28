# URL Shortener with PocketBase - Makefile
# Simplified Makefile with essential commands only

.PHONY: help build run stop clean logs test setup deps fmt

# Default target
help: ## Show this help message
	@echo "URL Shortener with PocketBase - Available commands:"
	@echo ""
	@echo "🐳 Docker Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E "^(build:|run:|stop:|clean:|logs:)" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "💻 Development Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E "(local-|test|fmt:|setup:|deps:)" | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[32m%-15s\033[0m %s\n", $$1, $$2}'

# =================================================================
# DOCKER COMMANDS
# =================================================================

# Detect container tool (docker or podman)
CONTAINER_TOOL := $(shell which docker 2>/dev/null || which podman 2>/dev/null)
COMPOSE_TOOL := $(shell which docker-compose 2>/dev/null || which podman-compose 2>/dev/null)

build: ## Build the container image
	@echo "Building container image using $(CONTAINER_TOOL)..."
	$(CONTAINER_TOOL) build -t url-shortener-pocketbase .

run: ## Run the application with container compose
	@echo "Starting application using $(COMPOSE_TOOL)..."
	$(COMPOSE_TOOL) up -d

stop: ## Stop the running application
	@echo "Stopping application..."
	$(COMPOSE_TOOL) down

logs: ## View application logs
	$(COMPOSE_TOOL) logs -f

clean: ## Clean up container resources
	@echo "Cleaning up container resources..."
	$(COMPOSE_TOOL) down -v
	$(CONTAINER_TOOL) system prune -f

# =================================================================
# LOCAL DEVELOPMENT COMMANDS
# =================================================================

local-build: ## Build the application locally
	@echo "Building application locally..."
	CGO_ENABLED=0 go build -o url-shortener .

local-run: ## Run the application locally
	@echo "Starting application locally..."
	./url-shortener serve --http=0.0.0.0:8090

local-stop: ## Stop the locally running application
	@echo "Stopping local application..."
	@pkill -f url-shortener || echo "No local application running"

# =================================================================
# DEVELOPMENT COMMANDS
# =================================================================

test: ## Run all tests
	@echo "Running all tests..."
	go test ./...

test-verbose: ## Run all tests with verbose output
	@echo "Running all tests (verbose)..."
	go test -v ./...

fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

deps: ## Check and install dependencies
	@echo "Checking dependencies..."
	@if ! which docker >/dev/null 2>&1 && ! which podman >/dev/null 2>&1; then \
		echo "Docker or Podman is required but not installed."; \
		exit 1; \
	fi
	@if ! which docker-compose >/dev/null 2>&1 && ! which podman-compose >/dev/null 2>&1; then \
		echo "Docker Compose or Podman Compose is required but not installed."; \
		exit 1; \
	fi
	@echo "Installing Go dependencies..."
	go mod download

setup: ## Initial setup
	@echo "Setting up project..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
	fi
	@echo "Installing dependencies..."
	@$(MAKE) deps
	@echo "Building application..."
	@$(MAKE) build
	@echo "Setup complete! Run 'make run' to start the application"

# Default make target
.DEFAULT_GOAL := help
