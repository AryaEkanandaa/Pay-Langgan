include .env
export

# Variables
GO ?= go
GOOSE ?= goose
MIGRATIONS_DIR ?= migrations
DB_DSN ?= host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)

.PHONY: run migrate-up migrate-down migrate-status test tidy

run:
	$(GO) run ./cmd/api

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status

test:
	$(GO) test ./... -v

tidy:
	$(GO) mod tidy
	$(GO) mod verify

build:
	$(GO) build -o bin/api ./cmd/api
	$(GO) build -o bin/worker ./cmd/worker

.PHONY: docker-up docker-down

test-api:
	newman run postman/PAY-LANGGAN.postman_collection.json -e postman/PAY-LANGGAN.local.postman_environment.json

.PHONY: docker-up docker-down test-api

docker-up:
	docker compose up -d

docker-down:
	docker compose down
