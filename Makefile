# TaskFlow API Makefile

# Load environment variables
include .env
export

# Application settings
APP_NAME=taskflow-api
BUILD_DIR=./build
MAIN_FILE=./cmd/main.go

# Docker settings
DOCKER_IMAGE=taskflow-api
DOCKER_TAG=latest

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
	@printf "\n"
	@printf "\033[0;36m╔══════════════════════════════════════════════════════════════╗\033[0m\n"
	@printf "\033[0;36m║\033[1;37m                    TaskFlow API - Available Commands          \033[0;36m║\033[0m\n"
	@printf "\033[0;36m╚══════════════════════════════════════════════════════════════╝\033[0m\n"
	@printf "\n"
	@printf "\033[1;33m🚀 Development Commands:\033[0m\n"
	@printf "  \033[0;32msetup\033[0m                    Complete development environment setup\n"
	@printf "  \033[0;32mrun\033[0m                      Run API locally\n"
	@printf "  \033[0;32mdev\033[0m                      Start database and run API locally\n"
	@printf "  \033[0;32mbuild\033[0m                    Build the application\n"
	@printf "  \033[0;32mtest\033[0m                     Run all tests\n"
	@printf "  \033[0;32mtest-coverage\033[0m            Run tests with coverage\n"
	@printf "  \033[0;32mlint\033[0m                     Run linter\n"
	@printf "  \033[0;32mformat\033[0m                   Format Go code\n"
	@printf "  \033[0;32mmod-tidy\033[0m                 Tidy go modules\n"
	@printf "  \033[0;32mclean\033[0m                    Clean build artifacts\n"
	@printf "\n"
	@printf "\033[0;34m🐳 Docker Commands:\033[0m\n"
	@printf "  \033[0;32mdocker-up\033[0m                Start all services (DB + API)\n"
	@printf "  \033[0;32mdocker-down\033[0m              Stop and remove all containers\n"
	@printf "  \033[0;32mdocker-restart\033[0m           Restart all services\n"
	@printf "  \033[0;32mdocker-build\033[0m             Build Docker image\n"
	@printf "  \033[0;32mdocker-logs\033[0m              Show logs for all services\n"
	@printf "  \033[0;32mlogs\033[0m                     Show logs for all services (alias)\n"
	@printf "  \033[0;32mdocker-clean\033[0m             Clean up Docker resources (images, volumes, networks)\n"
	@printf "  \033[0;32mstart\033[0m                    Quick start (alias for docker-up)\n"
	@printf "  \033[0;32mstop\033[0m                     Quick stop (alias for docker-down)\n"
	@printf "\n"
	@printf "\033[0;35m🐘 Database Commands:\033[0m\n"
	@printf "  \033[0;32mdb-up\033[0m                    Start only the database\n"
	@printf "  \033[0;32mdb-down\033[0m                  Stop the database\n"
	@printf "  \033[0;32mdb-shell\033[0m                 Access database shell\n"
	@printf "  \033[0;32mdb-backup\033[0m                Create database backup\n"
	@printf "  \033[0;32mdb-restore\033[0m               Restore database from backup (requires BACKUP_FILE variable)\n"
	@printf "  \033[0;32mdb-reset\033[0m                 \033[0;31mReset database (DANGEROUS - removes all data)\033[0m\n"
	@printf "\n"
	@printf "\033[0;36m🔄 Migration Commands:\033[0m\n"
	@printf "  \033[0;32minstall-goose\033[0m            Install Goose migration tool\n"
	@printf "  \033[0;32mmigrate\033[0m                  Run all pending migrations\n"
	@printf "  \033[0;32mmigrate-up\033[0m               Apply all pending migrations\n"
	@printf "  \033[0;32mmigrate-down\033[0m             Rollback last migration\n"
	@printf "  \033[0;32mmigrate-status\033[0m           Show detailed migration status\n"
	@printf "  \033[0;32mmigrate-create\033[0m           Create a new migration file (requires NAME variable)\n"
	@printf "  \033[0;32mmigrate-to\033[0m               Migrate to specific version (requires VERSION variable)\n"
	@printf "  \033[0;32mmigrate-reset\033[0m            \033[0;31mReset all migrations (DANGEROUS - removes all data)\033[0m\n"
	@printf "\n"
	@printf "\033[1;33m🛠️  Utility Commands:\033[0m\n"
	@printf "  \033[0;32mhelp\033[0m                     Show this help message\n"
	@printf "  \033[0;32mhealth\033[0m                   Check API health\n"
	@printf "  \033[0;32minfo\033[0m                     Show project information\n"
	@printf "  \033[0;32menv-example\033[0m              Copy .env.example to .env\n"
	@printf "\n"
	@printf "\033[1;37m📝 Usage Examples:\033[0m\n"
	@printf "  \033[0;36mmake setup\033[0m               # Initial project setup\n"
	@printf "  \033[0;36mmake start\033[0m               # Start everything with Docker\n"
	@printf "  \033[0;36mmake dev\033[0m                 # Development mode (DB in Docker, API local)\n"
	@printf "  \033[0;36mmake migrate-create NAME=add_users\033[0m # Create new migration\n"
	@printf "  \033[0;36mmake db-backup\033[0m           # Backup database\n"
	@printf "\n"
	@printf "\033[1;37m🌐 URLs:\033[0m\n"
	@printf "  API:         \033[0;36mhttp://localhost:$(PORT)\033[0m\n"
	@printf "  Health:      \033[0;36mhttp://localhost:$(PORT)/health\033[0m\n"
	@printf "  Database:    \033[0;36mlocalhost:$(DB_PORT)\033[0m\n"
	@printf "\n"

## Development Commands
.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@go mod tidy
	@echo "✅ Setup complete! Please update .env with your settings."

.PHONY: run
run: ## Run the application
	@echo "🚀 Starting TaskFlow API..."
	@go run $(MAIN_FILE)

.PHONY: build
build: ## Build the application
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

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

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

.PHONY: lint
lint: ## Run linter
	@echo "🔍 Running linter..."
	@golangci-lint run

.PHONY: format
format: ## Format code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@goimports -w .

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "📦 Tidying go modules..."
	@go mod tidy
	@go mod verify

## Docker Commands
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

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

## Database Commands
.PHONY: db-up
db-up: ## Start only the database
	@echo "🐘 Starting database..."
	@docker compose up -d postgres
	@echo "✅ Database started"

.PHONY: db-down
db-down: ## Stop the database
	@echo "🐘 Stopping database..."
	@docker compose stop postgres
	@echo "✅ Database stopped"

.PHONY: db-shell
db-shell: ## Connect to database shell
	@echo "🐘 Connecting to database..."
	@docker exec -it $(DB_CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-backup
db-backup: ## Backup database
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@docker exec $(DB_CONTAINER_NAME) pg_dump -U $(DB_USER) $(DB_NAME) > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created in backups/"

.PHONY: db-restore
db-restore: ## Restore database (requires BACKUP_FILE variable)
	@echo "📥 Restoring database from $(BACKUP_FILE)..."
	@docker exec -i $(DB_CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME) < $(BACKUP_FILE)
	@echo "✅ Database restored"

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "⚠️  WARNING: This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "\n🗑️  Resetting database..."; \
		docker compose down -v; \
		docker compose up -d postgres; \
		echo "✅ Database reset complete"; \
	else \
		echo "\n❌ Database reset cancelled"; \
	fi

## Migration Commands (Using Goose)
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

## Environment Commands
.PHONY: env-example
env-example: ## Copy .env.example to .env
	@cp .env.example .env
	@echo "✅ .env file created from .env.example"

## Health Check
.PHONY: health
health: ## Check API health
	@echo "🏥 Checking API health..."
	@curl -f http://localhost:$(PORT)/health || echo "❌ API is not responding"

## Quick Start
.PHONY: start
start: docker-up ## Quick start (alias for docker-up)

.PHONY: stop
stop: docker-down ## Quick stop (alias for docker-down)

.PHONY: dev
dev: db-up run ## Start database and run API locally

## Info
.PHONY: info
info: ## Show project information
	@echo "📋 TaskFlow API Information"
	@echo "   App Name: $(APP_NAME)"
	@echo "   Port: $(PORT)"
	@echo "   Environment: $(ENV)"
	@echo "   Database: $(DB_NAME)@$(DB_HOST):$(DB_PORT)"
	@echo "   API URL: http://localhost:$(PORT)"
	@echo "   Health Check: http://localhost:$(PORT)/health"