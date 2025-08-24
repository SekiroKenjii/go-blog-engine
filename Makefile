# Makefile for Go Blog Engine
# Use this file to manage common development tasks

# =============================================================================
# VARIABLES
# =============================================================================

# Project information
PROJECT_NAME := go-blog-engine
BINARY_NAME := server
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
BUILD_DIR := ./bin
CMD_DIR := ./cmd/server
DOC_DIR := ./docs
SQLC_DIR := ./internal/db/sqlc
SQL_SCHEMA_DIR := $(SQLC_DIR)/schemas
SQL_QUERY_DIR := $(SQLC_DIR)/queries
SCRIPTS_DIR := ./scripts

# Environment and configuration
ENV := develop
CONFIG_FILE_PATH := ./config/env/$(ENV).yml
DB_DRIVER := postgres

# Docker configuration
DOCKER_COMPOSE_FILE := ./environments/docker/docker-compose.yml
DOCKER_PROJECT_NAME := go-blog-engine

# Go build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

define read_db_conf
	$(shell yq '.$(DB_DRIVER).$(1)' $(CONFIG_FILE_PATH))
endef

# Goose environment variables
GOOSE_MIGRATION_DIR := $(SQL_SCHEMA_DIR)
GOOSE_DB_USER := $(firstword $(subst :, ,$(call read_db_conf,user)))
GOOSE_DB_PASSWORD := $(firstword $(subst :, ,$(call read_db_conf,password)))
GOOSE_DB_HOST := $(firstword $(subst :, ,$(call read_db_conf,host)))
GOOSE_DB_PORT := $(firstword $(subst :, ,$(call read_db_conf,port)))
GOOSE_DB_NAME := $(firstword $(subst :, ,$(call read_db_conf,name)))
GOOSE_DBSTRING := $(DB_DRIVER)://$(GOOSE_DB_USER):$(GOOSE_DB_PASSWORD)@$(GOOSE_DB_HOST):$(GOOSE_DB_PORT)/$(GOOSE_DB_NAME)

# =============================================================================
# DEFAULT TARGET
# =============================================================================

.DEFAULT_GOAL := help

# =============================================================================
# HELP
# =============================================================================

.PHONY: help
help: ## Display this help message
	@echo "$(PROJECT_NAME) - Development Commands"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# =============================================================================
# BUILD & RUN
# =============================================================================

.PHONY: build
build: ## Build the application binary
	@echo "Building $(PROJECT_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

.PHONY: run
run: ## Run the application (alias for server)
	@$(MAKE) server

.PHONY: server
server: ## Run the development server
	@echo "Starting development server (env: $(ENV))..."
	@go run $(CMD_DIR)/*.go $(env)

# =============================================================================
# DEVELOPMENT
# =============================================================================

.PHONY: dev
dev: ## Start development environment (docker + server)
	@echo "Starting development environment..."
	@$(MAKE) docker-up
	@sleep 3
	@$(MAKE) migrate-up
	@$(MAKE) server

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# =============================================================================
# TESTING & QUALITY
# =============================================================================

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run --config .golangci.yml

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w $(GO_FILES)

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: check
check: fmt vet lint test ## Run all quality checks

# =============================================================================
# DATABASE MIGRATIONS
# =============================================================================

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then echo "Error: name parameter required. Usage: make migrate-create name=migration_name"; exit 1; fi
	@echo "Creating migration: $(name)"
	@goose create -dir $(SQL_SCHEMA_DIR) $(name) sql

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) up

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) down

.PHONY: migrate-reset
migrate-reset: ## Reset database (WARNING: This will drop all data)
	@echo "WARNING: This will reset the database and drop all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "\nResetting database..."; \
		GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) reset; \
	else \
		echo "\nAborted."; \
	fi

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "Migration status:"
	@GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) status

# =============================================================================
# DOCKER
# =============================================================================

.PHONY: docker-up
docker-up: ## Start Docker services
	@echo "Starting Docker services..."
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) up -d

.PHONY: docker-build
docker-build: ## Build and start Docker services
	@echo "Building and starting Docker services..."
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) up --build -d

.PHONY: docker-down
docker-down: ## Stop and remove Docker services
	@echo "Stopping Docker services..."
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) down --volumes

.PHONY: docker-stop
docker-stop: ## Stop Docker services
	@echo "Stopping Docker services..."
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) stop

.PHONY: docker-start
docker-start: ## Start existing Docker services
	@echo "Starting Docker services..."
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) start

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) logs -f

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	@docker compose --file $(DOCKER_COMPOSE_FILE) --project-name $(DOCKER_PROJECT_NAME) down --volumes --remove-orphans
	@docker system prune -f

# =============================================================================
# CODE GENERATION
# =============================================================================

.PHONY: generate
generate: sqlc-gen swag-gen ## Run all code generation

.PHONY: sqlc-gen
sqlc-gen: ## Generate SQLC code
	@echo "Generating SQLC code..."
	@sqlc generate -f '$(SQLC_DIR)/sqlc.yml'

.PHONY: swag-init
swag-init: ## Initialize Swagger documentation
	@echo "Initializing Swagger documentation..."
	@swag init -d $(CMD_DIR) $(g)

.PHONY: swag-gen
swag-gen: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init -d $(dir) -g handler.go -o $(DOC_DIR) --parseDependency

# =============================================================================
# UTILITY
# =============================================================================

.PHONY: setup
setup: deps generate ## Setup development environment
	@echo "Setting up development environment..."
	@$(MAKE) docker-up
	@sleep 3
	@$(MAKE) migrate-up

.PHONY: reset
reset: docker-clean migrate-reset setup ## Reset entire development environment

.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Prevent make from interpreting migration names as targets
%:
	@:
