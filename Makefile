ENV = develop
CONFIG_FILE_PATH = ./configs/env/$(ENV).yml
DB_DRIVER = postgres

define read_db_conf
	$(shell yq '.$(DB_DRIVER).$(1)' $(CONFIG_FILE_PATH))
endef

MAIN_DIR = ./cmd/server
DOC_DIR = ./docs
SQLC_DIR = ./internal/db/sqlc
SQL_SCHEMA_DIR = $(SQLC_DIR)/schemas
SQL_QUERY_DIR = $(SQLC_DIR)/queries

.PHONY: server
server:
	@go run ./cmd/server/*.go $(filter-out $@,$(MAKECMDGOALS))

.PHONY: docker-up
docker-up:
	@docker compose --file 'environments/docker/docker-compose$(filter-out $@,$(MAKECMDGOALS)).yml' up -d

.PHONY: docker-build
docker-build:
	@docker compose --file 'environments/docker/docker-compose$(filter-out $@,$(MAKECMDGOALS)).yml' up --build -d

.PHONY: docker-down
docker-down:
	@docker compose --file 'environments/docker/docker-compose$(filter-out $@,$(MAKECMDGOALS)).yml' down --volumes

.PHONY: docker-stop
docker-stop:
	@docker compose --file 'environments/docker/docker-compose$(filter-out $@,$(MAKECMDGOALS)).yml' --project-name 'go-blog-engine' stop

.PHONY: docker-start
docker-start:
	@docker compose --file 'environments/docker/docker-compose$(filter-out $@,$(MAKECMDGOALS)).yml' --project-name 'go-blog-engine' start

.PHONY: swagger-init
swagger-init:
	@swag init -d $(MAIN_DIR) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: swagger-gen
swagger-gen:
	@swag init -d $(filter-out $@,$(MAKECMDGOALS)) -g handler.go -o $(DOC_DIR) --parseDependency

.PHONY: sqlc-gen
sqlc-gen:
	@sqlc generate -f '$(SQLC_DIR)/sqlc.yml'
