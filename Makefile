-include .env
export

MIGRATIONS_DIR ?= db/migrations
DB_DRIVER ?= postgres
DB_DSN ?= $(DATABASE_URL)

TAILWIND_INPUT ?= assets/css/app.css
TAILWIND_OUTPUT ?= static/css/app.css

.PHONY: migrate-create migrate-up migrate-down migrate-status sqlc-generate sqlc-vet sqlc-check docker-up docker-down templ-generate templ-watch tailwind tailwind-watch

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

templ-generate:
	templ generate

templ-watch:
	templ generate --watch

tailwind:
	tailwindcss -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --minify

tailwind-watch:
	tailwindcss -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --watch
