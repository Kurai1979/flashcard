-include .env
export

MIGRATIONS_DIR ?= db/migrations
DB_DRIVER ?= postgres
DB_DSN ?= $(DATABASE_URL)

.PHONY: migrate-create migrate-up migrate-down migrate-status sqlc-generate sqlc-vet sqlc-check docker-up docker-down

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_flashcards_table" && exit 1)
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" status

sqlc-generate:
	sqlc generate

sqlc-vet:
	sqlc vet

sqlc-check:
	sqlc generate
	go test ./...
