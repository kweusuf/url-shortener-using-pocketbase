# URL Shortener with PocketBase - Makefile
# This Makefile provides convenient commands for building, running, and managing the application

.PHONY: help build run stop clean logs dev prod test setup deps status

# Default target
help: ## Show this help message
	@echo "URL Shortener with PocketBase - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the Docker image
	@echo "Building Docker image..."
	docker build -t url-shortener-pocketbase .

build-no-cache: ## Build the Docker image without using cache
	@echo "Building Docker image (no cache)..."
	docker build --no-cache -t url-shortener-pocketbase .

# Run targets
run: ## Run the application in development mode
	@echo "Starting application..."
	docker-compose up -d

run-detached: ## Run the application in background
	@echo "Starting application in background..."
	docker run -d \
		--name url-shortener-app \
		-p 8090:8090 \
		-v ./pb_data:/app/pb_data \
		--restart unless-stopped \
		url-shortener-pocketbase

run-interactive: ## Run the application in foreground (for debugging)
	@echo "Starting application in foreground..."
	docker run \
		--name url-shortener-app \
		-p 8090:8090 \
		-v ./pb_data:/app/pb_data \
		url-shortener-pocketbase

# Stop targets
stop: ## Stop the running application
	@echo "Stopping application..."
	docker-compose down

stop-container: ## Stop and remove the Docker container
	@echo "Stopping and removing container..."
	docker stop url-shortener-app || true
	docker rm url-shortener-app || true

# Development targets
dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose up -d
	@echo "Application is running at http://localhost:8090"
	@echo "PocketBase admin panel at http://localhost:8090/_/"

dev-logs: ## View development logs
	docker-compose logs -f

# Production targets
prod: ## Deploy in production mode
	@echo "Deploying in production mode..."
	docker-compose -f docker-compose.yml up -d --build
	@echo "Production deployment complete"

prod-build: ## Build for production
	@echo "Building production image..."
	docker build -t url-shortener-pocketbase:prod .

# Clean targets
clean: ## Remove containers, networks, and volumes
	@echo "Cleaning up Docker resources..."
	docker-compose down -v
	docker system prune -f

clean-all: ## Remove everything including images
	@echo "Performing deep clean..."
	docker-compose down -v --rmi all
	docker system prune -af --volumes

# Log targets
logs: ## View application logs
	docker-compose logs -f

logs-container: ## View container logs
	docker logs -f url-shortener-app

# Status targets
status: ## Show status of running containers
	@echo "Container status:"
	docker ps --filter "name=url-shortener" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

ps: ## Show all Docker processes
	docker ps -a --filter "name=url-shortener"

# Utility targets
restart: ## Restart the application
	@echo "Restarting application..."
	docker-compose restart

shell: ## Access container shell
	docker exec -it url-shortener-app /bin/sh

# Database targets
db-backup: ## Backup PocketBase data
	@echo "Creating database backup..."
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker run --rm -v url-shortener_pb_data:/data -v $$(pwd):/backup alpine tar czf /backup/pb_data_backup_$${timestamp}.tar.gz -C /data .

db-restore: ## Restore PocketBase data (requires BACKUP_FILE variable)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Error: Please set BACKUP_FILE variable"; \
		echo "Usage: make db-restore BACKUP_FILE=pb_data_backup_20231201_120000.tar.gz"; \
		exit 1; \
	fi
	@echo "Restoring from $(BACKUP_FILE)..."
	docker run --rm -v url-shortener_pb_data:/data -v $$(pwd):/backup alpine sh -c "cd /data && tar xzf /backup/$(BACKUP_FILE)"

# Health check
health: ## Check application health
	@echo "Checking application health..."
	@curl -s http://localhost:8090/api/system/health | jq . 2>/dev/null || echo "Application not responding"

health-detailed: ## Check detailed application health
	@echo "Checking detailed application health..."
	@curl -s http://localhost:8090/api/system/health/detailed | jq . 2>/dev/null || echo "Application not responding"

health-ready: ## Check application readiness (for Docker/Kubernetes)
	@echo "Checking application readiness..."
	@curl -s http://localhost:8090/api/system/ready | jq . 2>/dev/null || echo "Application not ready"

health-live: ## Check application liveness (for Docker/Kubernetes)
	@echo "Checking application liveness..."
	@curl -s http://localhost:8090/api/system/live | jq . 2>/dev/null || echo "Application not alive"

metrics: ## Get application metrics
	@echo "Getting application metrics..."
	@curl -s http://localhost:8090/api/system/metrics | jq . 2>/dev/null || echo "Unable to get metrics"

# Setup targets
setup: ## Initial setup and build
	@echo "Setting up project..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
	fi
	@echo "Building and starting application..."
	$(MAKE) build
	$(MAKE) run

# Dependencies
deps: ## Install development dependencies (if needed)
	@echo "Checking dependencies..."
	@which docker >/dev/null 2>&1 || (echo "Docker is required but not installed." && exit 1)
	@which docker-compose >/dev/null 2>&1 || (echo "Docker Compose is required but not installed." && exit 1)
	@echo "All dependencies are installed."

# Test targets
test: ## Run all tests
	@echo "Running all tests..."
	go test ./...

test-verbose: ## Run all tests with verbose output
	@echo "Running all tests (verbose)..."
	go test -v ./...

test-coverage: ## Run tests with coverage analysis
	@echo "Running tests with coverage..."
	go test -cover ./...

test-generator: ## Run generator utility tests
	@echo "Running generator tests..."
	go test -v ./pkg/utils/generator/...

test-url: ## Run URL utility tests
	@echo "Running URL utility tests..."
	go test -v ./pkg/utils/url/...

test-db: ## Run database utility tests
	@echo "Running database utility tests..."
	go test -v ./pkg/utils/db/...

test-ws: ## Run WebSocket utility tests
	@echo "Running WebSocket utility tests..."
	go test -v ./pkg/utils/ws/...

test-transport: ## Run transport layer tests
	@echo "Running transport layer tests..."
	go test -v ./transport/ws/...

test-monitoring: ## Run monitoring utility tests
	@echo "Running monitoring utility tests..."
	go test -v ./pkg/utils/monitoring/...

test-health: ## Test if the application is healthy
	@echo "Testing application health..."
	@sleep 2
	@if curl -s http://localhost:8090/api/hello > /dev/null; then \
		echo "✅ Application is healthy"; \
	else \
		echo "❌ Application is not responding"; \
		exit 1; \
	fi

test-coverage-html: ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test -bench=. ./...

test-bench-generator: ## Run generator benchmarks
	@echo "Running generator benchmarks..."
	go test -bench=. ./pkg/utils/generator/...

test-bench-url: ## Run URL utility benchmarks
	@echo "Running URL utility benchmarks..."
	go test -bench=. ./pkg/utils/url/...

test-bench-ws: ## Run WebSocket benchmarks
	@echo "Running WebSocket benchmarks..."
	go test -bench=. ./pkg/utils/ws/...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -race ./...

test-clean: ## Clean test artifacts
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html

# Update targets
update: ## Update the application (build new image and restart)
	@echo "Updating application..."
	$(MAKE) build
	docker-compose up -d --build

# Info targets
info: ## Show information about the application
	@echo "=== URL Shortener with PocketBase ==="
	@echo "Docker Image: url-shortener-pocketbase"
	@echo "Container Name: url-shortener-app"
	@echo "External Port: 8090"
	@echo "Data Directory: ./pb_data"
	@echo ""
	@echo "Useful URLs:"
	@echo "  API: http://localhost:8090/api/hello"
	@echo "  Admin: http://localhost:8090/_/"
	@echo "  Health: http://localhost:8090/api/hello"

# Default make target
.DEFAULT_GOAL := help
