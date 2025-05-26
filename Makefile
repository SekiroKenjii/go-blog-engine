# Makefile for Go Blog Engine

# Define environment variables
ENV = develop
CONFIG_FILE_PATH = ./config/env/$(ENV).yml
DB_DRIVER = postgres
MAIN_DIR = ./cmd/server
DOC_DIR = ./docs
SQLC_DIR = ./internal/db/sqlc
SQL_SCHEMA_DIR = $(SQLC_DIR)/schemas
SQL_QUERY_DIR = $(SQLC_DIR)/queries

# Define common functions
define read_db_conf
	$(shell yq '.$(DB_DRIVER).$(1)' $(CONFIG_FILE_PATH))
endef

# Goose environment variables
GOOSE_MIGRATION_DIR = $(SQL_SCHEMA_DIR)
GOOSE_DB_USER = $(firstword $(subst :, ,$(call read_db_conf,user)))
GOOSE_DB_PASSWORD = $(firstword $(subst :, ,$(call read_db_conf,password)))
GOOSE_DB_HOST = $(firstword $(subst :, ,$(call read_db_conf,host)))
GOOSE_DB_PORT = $(firstword $(subst :, ,$(call read_db_conf,port)))
GOOSE_DB_NAME = $(firstword $(subst :, ,$(call read_db_conf,name)))
GOOSE_DBSTRING = $(DB_DRIVER)://$(GOOSE_DB_USER):$(GOOSE_DB_PASSWORD)@$(GOOSE_DB_HOST):$(GOOSE_DB_PORT)/$(GOOSE_DB_NAME)

# Goose migration commands
.PHONY: migrate-create
migrate-create:
	@goose create -dir $(SQL_SCHEMA_DIR) $(filter-out $@,$(MAKECMDGOALS)) sql

.PHONY: migrate-up
migrate-up:
	@GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) up

.PHONY: migrate-down
migrate-down:
	@GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) down

.PHONY: migrate-reset
migrate-reset:
	@GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir=$(GOOSE_MIGRATION_DIR) reset

# Docker commands
.PHONY: docker-up
docker-up:
	@docker compose --file environments/docker/docker-compose.yml up -d

.PHONY: docker-build
docker-build:
	@docker compose --file environments/docker/docker-compose.yml up --build -d

.PHONY: docker-down
docker-down:
	@docker compose --file environments/docker/docker-compose.yml down --volumes

.PHONY: docker-stop
docker-stop:
	@docker compose --file environments/docker/docker-compose.yml --project-name 'go-blog-engine' stop

.PHONY: docker-start
docker-start:
	@docker compose --file environments/docker/docker-compose.yml --project-name 'go-blog-engine' start

# Go commands
.PHONY: server
server:
	@go run ./cmd/server/*.go $(env)

.PHONY: swag-init
swag-init:
	@swag init -d $(MAIN_DIR) $(g)

.PHONY: swag-gen
swag-gen:
	@swag init -d $(dir) -g handler.go -o $(DOC_DIR) --parseDependency

.PHONY: sqlc-gen
sqlc-gen:
	@sqlc generate -f '$(SQLC_DIR)/sqlc.yml'
