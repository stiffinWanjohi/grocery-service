include .env
export

MIGRATION_DIR=migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: migrate-create migrate-up migrate-down migrate-force docker-build docker-up docker-down docker-logs

# Existing migration commands
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $${name}; \
	echo "Migration files created in $(MIGRATION_DIR):"; \
	find $(MIGRATION_DIR) -name "*_$${name}.*.sql"

migrate-up:
	@if [ -z "$(steps)" ]; then \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up; \
	else \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up $(steps); \
	fi

migrate-down:
	@if [ -z "$(steps)" ]; then \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down; \
	else \
		migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down $(steps); \
	fi

migrate-force:
	@read -p "Enter version: " version; \
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $$version

# Database commands
db-create:
	PGPASSWORD=$(DB_PASSWORD) createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME)

db-drop:
	PGPASSWORD=$(DB_PASSWORD) dropdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME)

db-reset: db-drop db-create migrate-up

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-migrate:
	docker-compose exec api ./migrate up

docker-reset: docker-down
	docker-compose down -v
	docker-compose up -d

# Installation commands
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Development commands
dev:
	docker-compose up --build

.DEFAULT_GOAL := help
help:
	@echo "Available commands:"
	@echo "  make migrate-create    - Create a new migration file"
	@echo "  make migrate-up        - Apply all or N up migrations"
	@echo "  make migrate-down      - Roll back all or N migrations"
	@echo "  make migrate-force     - Force set version"
	@echo "  make db-create        - Create database"
	@echo "  make db-drop         - Drop database"
	@echo "  make db-reset        - Reset database"
	@echo "  make install-migrate  - Install migration tool"
	@echo "Docker commands:"
	@echo "  make docker-build    - Build Docker images"
	@echo "  make docker-up       - Start Docker containers"
	@echo "  make docker-down     - Stop Docker containers"
	@echo "  make docker-logs     - View Docker logs"
	@echo "  make docker-migrate  - Run migrations in Docker"
	@echo "  make docker-reset    - Reset Docker environment"
	@echo "  make dev            - Start development environment"