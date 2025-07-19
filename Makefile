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

## Development Commands
.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@go mod tidy
	@echo "✅ Setup complete! Please update .env with your settings."

.PHONY: dev-setup
dev-setup: ## Complete development environment setup
	@echo "${BLUE}🔧 Setting up development environment...${NC}"
	@$(MAKE) check-deps
	@$(MAKE) env-check
	@$(MAKE) docker-down
	@$(MAKE) docker-build
	@$(MAKE) db-up
	@$(MAKE) migrate
	@echo "${GREEN}✅ Development environment ready!${NC}"

.PHONY: run
run: ## Run the application
	@echo "🚀 Starting TaskFlow API..."
	@go run $(MAIN_FILE)

.PHONY: dev-run
dev-run: ## Start full development environment (DB + API in containers)
	@echo "${BLUE}🚀 Starting development environment...${NC}"
	@$(MAKE) docker-down
	@$(MAKE) db-up
	@$(MAKE) docker-build-api
	@$(DOCKER_COMPOSE) up -d api
	@echo "${GREEN}✅ Development environment is running!${NC}"
	@$(MAKE) logs-api

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

.PHONY: build
build: ## Build the application
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

## Docker Commands
.PHONY: docker-build-api
docker-build-api: ## Build only API Docker image
	@echo "${BLUE}🔨 Building API Docker image...${NC}"
	@$(DOCKER_COMPOSE) build api
	@echo "${GREEN}✅ API Docker image built!${NC}"

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

## Enhanced Database Commands
.PHONY: db-wait
db-wait: ## Wait for database to be ready
	@echo "${YELLOW}⏳ Waiting for database connection...${NC}"
	@timeout 60 bash -c 'until docker exec ${DB_CONTAINER_NAME} pg_isready -U $(DB_USER); do sleep 1; done' || (echo "${RED}❌ Database failed to start in 60 seconds${NC}" && exit 1)
	@echo "${GREEN}✅ Database is ready!${NC}"

## Enhanced Logging Commands
.PHONY: logs-api
logs-api: ## Show API logs only
	@$(DOCKER_COMPOSE) logs -f api

.PHONY: logs-db
logs-db: ## Show database logs only
	@$(DOCKER_COMPOSE) logs -f db

## Enhanced Migration Commands
.PHONY: goose-check
goose-check: ## Check if Goose is available and database is accessible
	@command -v goose >/dev/null 2>&1 || (echo "${RED}❌ Goose is not installed or not in PATH${NC}" && exit 1)
	@if ! docker ps | grep -q ${DB_CONTAINER_NAME}; then \
		echo "${RED}❌ Database container not running. Start it with 'make db-up'${NC}"; \
		exit 1; \
	fi
	@if ! docker exec ${DB_CONTAINER_NAME} pg_isready -U $(DB_USER) >/dev/null 2>&1; then \
		echo "${RED}❌ Database is not accepting connections${NC}"; \
		exit 1; \
	fi

.PHONY: migrate-summary
migrate-summary: ## Show migration summary (counts)
	@echo "${BLUE}📈 Migration Summary:${NC}"
	@echo -n "${GREEN}✅ Total migration files: ${NC}"
	@ls migrations/*.sql | wc -l
	@echo -n "${YELLOW}📝 Applied migrations: ${NC}"
	@goose -dir migrations postgres "$(DB_URL)" status 2>&1 | grep -v "Applied At" | grep -v "=======" | grep -v "Pending" | grep " -- " | wc -l || echo "0"
	@echo -n "${RED}⏳ Pending migrations: ${NC}"
	@goose -dir migrations postgres "$(DB_URL)" status 2>&1 | grep "Pending" | wc -l || echo "0"

.PHONY: migrate-version
migrate-version: ## Show current migration version
	@echo "${BLUE}🔍 Current Migration Version:${NC}"
	@$(MAKE) goose-check
	@goose -dir migrations postgres "$(DB_URL)" version

.PHONY: migrate-sync
migrate-sync: ## Sync Goose with existing database state (mark all migrations as applied)
	@echo "${BLUE}🔄 Syncing Goose migration status with existing database...${NC}"
	@$(MAKE) goose-check
	@echo "${YELLOW}⚠️  This will mark all existing migrations as applied without running them${NC}"
	@read -p "Continue? Database appears to have all tables already. (y/N): " confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ] || (echo "Sync cancelled." && exit 1)
	@echo "${BLUE}📝 Marking migrations as applied...${NC}"
	@for migration in $$(ls migrations/*.sql | sort | xargs -n1 basename | sed 's/.sql//' | sed 's/_.*//'); do \
		echo "INSERT INTO goose_db_version (version_id, is_applied, tstamp) SELECT $$migration, true, NOW() WHERE NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version_id = $$migration);" | \
		docker exec -i ${DB_CONTAINER_NAME} psql -U $(DB_USER) -d $(DB_NAME); \
	done
	@echo "${GREEN}✅ Migration sync complete!${NC}"
	@$(MAKE) migrate-summary

## Utility Commands
.PHONY: env-check
env-check: ## Check if .env file exists
	@if [ ! -f .env ]; then \
		echo "${YELLOW}⚠️  .env file not found. Creating from .env.example...${NC}"; \
		cp .env.example .env; \
		echo "${GREEN}✅ .env file created. Please update it with your values.${NC}"; \
	fi

.PHONY: check-deps
check-deps: ## Check required dependencies
	@echo "${BLUE}🔍 Checking dependencies...${NC}"
	@command -v docker >/dev/null 2>&1 || (echo "${RED}❌ Docker is required but not installed${NC}" && exit 1)
	@command -v docker-compose >/dev/null 2>&1 || command -v docker compose >/dev/null 2>&1 || (echo "${RED}❌ Docker Compose is required but not installed${NC}" && exit 1)
	@command -v go >/dev/null 2>&1 || (echo "${RED}❌ Go is required but not installed${NC}" && exit 1)
	@echo "${GREEN}✅ All dependencies are available!${NC}"

## Legacy Commands (for backward compatibility)
.PHONY: up down db start_be
up: docker-up ## Legacy: Start all services
down: docker-down ## Legacy: Stop all services  
db: db-up ## Legacy: Start database
start_be: dev-run ## Legacy: Start backend services

# Default target
.DEFAULT_GOAL := help

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

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
	@docker compose logs -f

.PHONY: logs
logs: docker-logs ## Alias for docker-logs

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	@echo "🧹 Cleaning Docker resources..."
	@docker compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Docker cleanup complete"

.PHONY: db-up
db-up: ## Start only the database
	@echo "🐘 Starting database..."
	@docker compose up -d db
	@echo "✅ Database started"

.PHONY: db-down
db-down: ## Stop the database
	@echo "🐘 Stopping database..."
	@docker compose stop db
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
	@echo "📥 Restoring database from $(BACKUP_FILE)..."
	@docker exec -i ${DB_CONTAINER_NAME} psql -U $(DB_USER) -d $(DB_NAME) < $(BACKUP_FILE)
	@echo "✅ Database restored"

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "⚠️  WARNING: This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "\n🗑️  Resetting database..."; \
		docker compose down -v; \
		docker compose up -d db; \
		echo "✅ Database reset complete"; \
	else \
		echo "\n❌ Database reset cancelled"; \
	fi

.PHONY: install-goose
install-goose: ## Install Goose migration tool
	@echo "📦 Installing Goose..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "✅ Goose installed"

.PHONY: migrate
migrate: migrate-up ## Alias for migrate-up

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
	@echo "📝 Creating migration: $(NAME)"
	@goose -dir migrations create $(NAME) sql
	@echo "✅ Migration created"

.PHONY: migrate-to
migrate-to: ## Migrate to specific version (requires VERSION variable)
	@echo "🎯 Migrating to version: $(VERSION)"
	@goose -dir migrations postgres "$(DB_URL)" to $(VERSION)
	@echo "✅ Migration to version $(VERSION) complete"

.PHONY: migrate-reset
migrate-reset: ## Reset all migrations (WARNING: This will delete all data)
	@echo "⚠️  WARNING: This will delete all migration data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "\n🗑️  Resetting migrations..."; \
		goose -dir migrations postgres "$(DB_URL)" reset; \
		echo "✅ Migrations reset complete"; \
	else \
		echo "\n❌ Migration reset cancelled"; \
	fi

.PHONY: env-example
env-example: ## Copy .env.example to .env
	@cp .env.example .env
	@echo "✅ .env file created from .env.example"

.PHONY: health
health: ## Check API health
	@echo "🏥 Checking API health..."
	@curl -f http://localhost:$(PORT)/health || echo "❌ API is not responding"

.PHONY: start
start: docker-up ## Quick start (alias for docker-up)

.PHONY: stop
stop: docker-down ## Quick stop (alias for docker-down)

.PHONY: dev
dev: db-up run ## Start database and run API locally

.PHONY: status
status: ## Show status of all services
	@echo "📊 TaskFlow API - Service Status"
	@echo "════════════════════════════════════════════════════════════"
	@echo ""
	@echo "🐳 Docker Compose Services:"
	@docker compose ps --format "table {{.Name}}\t{{.State}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "❌ No services running"
	@echo ""
	@echo "🌐 Network Status:"
	@docker network ls --filter name=$(NETWORK_NAME) --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}" 2>/dev/null || echo "❌ No custom networks found"
	@echo ""
	@echo "📦 Container Details:"
	@echo "Database Container: ${DB_CONTAINER_NAME}"
	@echo "API Container: ${API_CONTAINER_NAME}"
	@echo "Migration Container: ${MIGRATION_CONTAINER_NAME}"
	@echo ""
	@echo "🔗 Service URLs:"
	@echo "  API:         http://localhost:$(PORT)"
	@echo "  Health:      http://localhost:$(PORT)/health"
	@echo "  Database:    localhost:$(DB_PORT)"
	@echo ""

.PHONY: info
info: ## Show project information
	@echo "📋 TaskFlow API Information"
	@echo "   App Name: $(APP_NAME)"
	@echo "   Port: $(PORT)"
	@echo "   Environment: $(ENV)"
	@echo "   Database: $(DB_NAME)@$(DB_HOST):$(DB_PORT)"
	@echo "   API URL: http://localhost:$(PORT)"
	@echo "   Health Check: http://localhost:$(PORT)/health"