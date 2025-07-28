# TaskFlow API Makefile

# Load environment variables
include .env
export

# Application settings
APP_NAME=taskflow-api
BUILD_DIR=./build
MAIN_FILE=./cmd/main.go
BINARY_NAME=taskflow-api
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Docker settings
DOCKER_IMAGE=taskflow-api
DOCKER_TAG=latest
DOCKER_COMPOSE := docker compose
GO_CMD := go
GO_BUILD := $(GO_CMD) build
GO_RUN := $(GO_CMD) run
GO_TEST := $(GO_CMD) test
GO_CLEAN := $(GO_CMD) clean
GO_MOD := $(GO_CMD) mod

# Database settings
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Colors for help output
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
PURPLE=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[1;37m
NC=\033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "${CYAN}TaskFlow API - Available Commands:${NC}"
	@echo ""
	@echo "${YELLOW}Development Commands:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | grep -E "(dev|run|test|lint|format|build|setup|clean|vet|mod)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2}' | sort
	@echo ""
	@echo "${YELLOW}Docker Commands:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | grep -E "(docker|up|down|logs|start|stop)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2}' | sort
	@echo ""
	@echo "${YELLOW}Database Commands:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | grep -E "(db|migrate)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2}' | sort
	@echo ""
	@echo "${YELLOW}Production Commands:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | grep -E "(prod|deploy|release)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2}' | sort
	@echo ""
	@echo "${YELLOW}Utility Commands:${NC}"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | grep -vE "(dev|run|test|lint|format|build|setup|clean|vet|mod|docker|up|down|logs|start|stop|db|migrate|prod|deploy|release)" | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2}' | sort

## Main Setup Commands
.PHONY: setup
setup: ## Start all containers (db, api, migrations), wait for DB readiness, and run migrations
	@echo "${BLUE}🚀 Setting up complete TaskFlow environment...${NC}"
	@$(MAKE) env-check
	@echo "${BLUE}📦 Building containers...${NC}"
	@$(DOCKER_COMPOSE) build
	@echo "${BLUE}🐳 Starting all services...${NC}"
	@$(DOCKER_COMPOSE) up -d db api
	@echo "${YELLOW}⏳ Waiting for database to be ready...${NC}"
	@$(MAKE) db-wait
	@echo "${BLUE}🔄 Running database migrations...${NC}"
	@$(MAKE) migrate-up
	@echo "${GREEN}✅ Setup complete! All services are running with migrations applied.${NC}"
	@$(MAKE) status

.PHONY: dev-run
dev-run: ## Start full development environment (DB + API in containers)
	@echo "${BLUE}🚀 Starting development environment...${NC}"
	@$(MAKE) docker-down
	@$(MAKE) db-up
	@$(MAKE) db-wait
	@$(MAKE) migrate-up
	@$(MAKE) run
	@echo "${GREEN}✅ Development environment is running!${NC}"

## Development Commands
.PHONY: run
run: ## Run the application
	@echo "🚀 Starting TaskFlow API..."
	@go run $(MAIN_FILE)

## Enhanced Testing Commands
.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "${BLUE}🧪 Running tests with race detection...${NC}"
	@$(GO_TEST) -race -v ./...

.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## Code Quality Commands
.PHONY: vet
vet: ## Run go vet
	@echo "${BLUE}🔍 Running go vet...${NC}"
	@$(GO_CMD) vet ./...

.PHONY: lint
lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "${BLUE}🔍 Running linter...${NC}"; \
		golangci-lint run; \
	else \
		echo "${YELLOW}⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"; \
	fi

.PHONY: format
format: ## Format code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "📦 Tidying go modules..."
	@go mod tidy
	@go mod verify

## Build Commands
.PHONY: build-linux
build-linux: ## Build Go binary for Linux
	@echo "${BLUE}🔨 Building Go binary for Linux...${NC}"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BINARY_PATH)-linux $(MAIN_FILE)
	@echo "${GREEN}✅ Linux binary built: $(BINARY_PATH)-linux${NC}"

.PHONY: rebuild-linux
rebuild-linux: clean build-linux ## Clean and build for Linux
	@echo "${GREEN}✅ Linux rebuild complete!${NC}"

.PHONY: build
build: ## Build the application
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: rebuild
rebuild: clean build ## Clean and build the application
	@echo "${GREEN}✅ Rebuild complete!${NC}"

## Docker Commands
.PHONY: docker-build-api
docker-build-api: ## Build only API Docker image
	@echo "${BLUE}🔨 Building API Docker image...${NC}"
	@$(DOCKER_COMPOSE) build api
	@echo "${GREEN}✅ API Docker image built!${NC}"

.PHONY: docker-build
docker-build: ## Build all Docker images using Docker Compose
	@echo "🐳 Building Docker images for all services..."
	@$(DOCKER_COMPOSE) build
	@echo "✅ Docker images built"

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild Docker images (force rebuild)
	@echo "🔨 Rebuilding Docker images..."
	@$(DOCKER_COMPOSE) build --no-cache
	@echo "✅ Docker images rebuilt"

## Migration Commands (Using Goose)
.PHONY: install-goose
install-goose: ## Install Goose migration tool
	@echo "📦 Installing Goose..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "✅ Goose installed"

.PHONY: migrate-up
migrate-up: ## Run all pending migrations
	@echo "🔄 Running migrations..."
	@goose -dir migrations postgres "$(DB_URL)" up
	@echo "✅ Migrations complete"

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "⬇️  Rolling back last migration..."
	@goose -dir migrations postgres "$(DB_URL)" down
	@echo "✅ Rollback complete"

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "📊 Migration status:"
	@goose -dir migrations postgres "$(DB_URL)" status

.PHONY: migrate-create
migrate-create: ## Create new migration (requires NAME variable)
	@if [ -z "$(NAME)" ]; then \
		echo "${RED}❌ Error: NAME variable is required${NC}"; \
		echo "${BLUE}💡 Usage: make migrate-create NAME=your_migration_name${NC}"; \
		echo "${BLUE}💡 Example: make migrate-create NAME=create_users_table${NC}"; \
		exit 1; \
	fi
	@echo "${BLUE}📝 Creating migration: $(NAME)${NC}"
	@goose -dir migrations create $(NAME) sql
	@echo "${GREEN}✅ Migration created successfully${NC}"

.PHONY: migrate-reset
migrate-reset: ## Reset all migrations (WARNING: This will delete all data)
	@echo "⚠️  WARNING: This will delete all migration data!"
	@echo "Are you sure? [y/N] " && read ans && [ "$$ans" = y ] || { echo "❌ Migration reset cancelled"; exit 1; } \
	&& echo "Resetting migrations..." \
	&& goose -dir migrations postgres "$(DB_URL)" reset \
	&& echo "✅ Database reset complete"

.PHONY: migrate-summary
migrate-summary: ## Show migration summary (counts)
	@echo "${BLUE}📈 Migration Summary:${NC}"
	@echo -n "${GREEN}✅ Total migration files: ${NC}"
	@find migrations -name '*.sql' | wc -l
	@echo -n "${YELLOW}📝 Applied migrations: ${NC}"
	@goose -dir migrations postgres "$(DB_URL)" status 2>&1 | grep -v "Applied At" | grep -v "=======" | grep -v "Pending" | grep -c " -- " || echo "0"
	@echo -n "${RED}⏳ Pending migrations: ${NC}"
	@goose -dir migrations postgres "$(DB_URL)" status 2>&1 | grep -c "Pending" || echo "0"

## Enhanced Database Commands
.PHONY: db-wait
db-wait: ## Wait for database to be ready
	@echo "${YELLOW}⏳ Waiting for database connection... (max 60 seconds)${NC}"
	@for i in $$(seq 1 60); do \
		if docker exec ${DB_CONTAINER_NAME} pg_isready -U $(DB_USER) >/dev/null 2>&1; then \
			echo "${GREEN}✅ Database is ready!${NC}"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "${RED}❌ Database failed to start in 60 seconds${NC}"; \
	exit 1

## Enhanced Logging Commands
.PHONY: logs-api
logs-api: ## Show API logs only
	@if [ "$$(docker ps -q -f name=api)" = "" ]; then \
		echo "❌ API container is not running."; \
	else \
		$(DOCKER_COMPOSE) logs -f api; \
	fi

.PHONY: logs-db
logs-db: ## Show database logs only
	@if [ "$$(docker ps -q -f name=db)" = "" ]; then \
		echo "❌ DB container is not running."; \
	else \
		$(DOCKER_COMPOSE) logs -f db; \
	fi

## Quick Access Commands
.PHONY: start
start: docker-up ## Quick start (alias for docker-up)

.PHONY: stop
stop: docker-down ## Quick stop (alias for docker-down)

.PHONY: logs
logs: docker-logs ## Show logs for all services

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

## Utility Commands
.PHONY: env-check
env-check: ## Check if .env file exists and required variables are set
	@if [ ! -f .env ]; then \
		echo "${YELLOW}⚠️  .env file not found. Creating from .env.example...${NC}"; \
		cp .env.example .env; \
		echo "${GREEN}✅ .env file created. Please update it with your values.${NC}"; \
	else \
		echo "${BLUE}ℹ️  .env file found.${NC}"; \
	fi

	@REQUIRED_VARS="DB_CONTAINER_NAME API_CONTAINER_NAME PORT DB_USER DB_PASSWORD DB_HOST DB_PORT DB_NAME DB_SSL_MODE" \
	MISSING=0; \
	for var in $$REQUIRED_VARS; do \
		eval "value=\$$$$var"; \
		if [ -z "$$value" ]; then \
			echo "${RED}❌ Environment variable '$$var' is not set. Please check your .env file.${NC}"; \
			MISSING=1; \
		fi; \
	done; \
	if [ $$MISSING -eq 1 ]; then \
		exit 1; \
	fi

.PHONY: check-deps
check-deps: ## Check required dependencies
	@echo "${BLUE}🔍 Checking dependencies...${NC}"
	@command -v docker >/dev/null 2>&1 || (echo "${RED}❌ Docker is required but not installed${NC}" && exit 1)
	@command -v docker >/dev/null 2>&1 && (docker compose version >/dev/null 2>&1 || docker-compose --version >/dev/null 2>&1) || (echo "${RED}❌ Docker Compose is required but not installed${NC}" && exit 1)
	@if command -v go >/dev/null 2>&1; then \
		echo "${GREEN}Go is available${NC}"; \
	elif test -x /usr/local/go/bin/go; then \
		echo "${YELLOW}⚠️  Go found at /usr/local/go/bin/go but not in PATH${NC}"; \
	else \
		echo "${RED}❌ Go is required but not installed${NC}" && exit 1; \
	fi
	@echo "Checking recommended dependencies..."
	@command -v psql >/dev/null 2>&1 || echo "${YELLOW}⚠️  psql (PostgreSQL client) is recommended but not installed${NC}"
	@command -v goose >/dev/null 2>&1 || echo "${YELLOW}⚠️  goose migration tool is recommended but not installed. Install with: go install github.com/pressly/goose/v3/cmd/goose@latest${NC}"
	@echo "${GREEN}✅ All required dependencies are available!${NC}"

# Default target
.DEFAULT_GOAL := help

.PHONY: docker-up
docker-up: ## Start Docker containers
	@echo "🐳 Starting Docker containers..."
	@docker compose up -d
	@echo "✅ Containers started"

.PHONY: docker-down
docker-down: ## Stop Docker containers
	@echo "🐳 Stopping Docker containers..."
	@docker compose down
	@echo "✅ Containers stopped"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker containers

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	@echo "🧹 Cleaning Docker resources..."
	@$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans
	@echo "✅ Docker cleanup complete"

.PHONY: db-up
db-up: ## Start only the database
	@echo "🐘 Starting database..."
	@$(DOCKER_COMPOSE) up -d db
	@echo "✅ Database started"

.PHONY: db-down
db-down: ## Stop the database
	@echo "🐘 Stopping database..."
	@$(DOCKER_COMPOSE) stop db
	@echo "✅ Database stopped"

.PHONY: db-shell
db-shell: ## Connect to database shell
	@echo "🐘 Connecting to database..."
	@docker exec -it ${DB_CONTAINER_NAME} psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-backup
db-backup: ## Backup database
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@docker exec ${DB_CONTAINER_NAME} pg_dump -U $(DB_USER) $(DB_NAME) > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created in backups/"

.PHONY: db-restore
db-restore: ## Restore database (requires BACKUP_FILE variable)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "${RED}❌ BACKUP_FILE variable is required${NC}"; \
		echo "${BLUE}💡 Usage: make db-restore BACKUP_FILE=path/to/backup.sql${NC}"; \
		exit 1; \
	fi
	@echo "📥 Restoring database from $(BACKUP_FILE)..."
	@docker exec -i ${DB_CONTAINER_NAME} psql -U $(DB_USER) -d $(DB_NAME) < $(BACKUP_FILE)
	@echo "✅ Database restored"

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "⚠️  WARNING: This will delete all data!"
	@bash -c 'read -p "Are you sure? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		echo "🗑️  Resetting database..."; \
		$(DOCKER_COMPOSE) down -v; \
		$(DOCKER_COMPOSE) up -d db; \
		echo "✅ Database reset complete"; \
	else \
		echo "❌ Database reset cancelled"; \
	fi'

.PHONY: health
health: ## Check API health
	@echo "🏥 Checking API health..."
	@curl -f http://localhost:$(PORT)/health || echo "❌ API is not responding"

.PHONY: status
status: ## Show status of all services
	@echo "📊 TaskFlow API - Service Status"
	@echo "════════════════════════════════════════════════════════════"
	@echo ""
	@echo "🐳 Docker Compose Services:"
	@$(DOCKER_COMPOSE) ps --format "table {{.Name}}\t{{.State}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "❌ No services running"
	@echo ""
	@echo "🔗 Service URLs:"
	@echo "  API:         http://localhost:$(PORT)"
	@echo "  Health:      http://localhost:$(PORT)/health"
	@echo "  Database:    localhost:$(DB_PORT)"
	@echo ""
	@echo "🔍 Health Checks:"
	@if docker ps --filter name=$(DB_CONTAINER_NAME) --filter status=running -q | grep -q .; then \
		echo "  Database: ✅ Running"; \
	else \
		echo "  Database: ❌ Not running"; \
	fi
	@if docker ps --filter name=$(API_CONTAINER_NAME) --filter status=running -q | grep -q .; then \
		echo "  API: ✅ Running"; \
	else \
		echo "  API: ❌ Not running"; \
	fi
	@echo ""
