# flashcard

Minimal Go project scaffold with:

- Goose migrations in `db/migrations`
- sqlc schema in `db/schema`
- sqlc queries in `db/query`
- Generated sqlc code in `internal/db`

## Folder layout

```text
cmd/server/           # app entrypoint
db/migrations/        # goose migrations
db/schema/            # schema source for sqlc
db/query/             # named sqlc queries
internal/db/          # generated sqlc code
```

## Install tools

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Migration commands

```bash
make migrate-create name=create_flashcards_table
make migrate-status DB_DRIVER=postgres DB_DSN="$DATABASE_URL"
make migrate-up DB_DRIVER=postgres DB_DSN="$DATABASE_URL"
make migrate-down DB_DRIVER=postgres DB_DSN="$DATABASE_URL"
```

## sqlc commands

```bash
make sqlc-generate
make sqlc-vet
make sqlc-check
```

