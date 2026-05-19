# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Tools

Install CLI tools before developing:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Common commands

```bash
# Database
make docker-up                              # start postgres container
make migrate-up                             # apply all pending migrations
make migrate-down                           # roll back last migration
make migrate-create name=create_foo_table   # create a new migration

# sqlc
make sqlc-generate   # regenerate internal/db/ from db/query/
make sqlc-vet        # lint queries against the database
make sqlc-check      # generate + run tests
```

`DATABASE_URL` is read from `.env` (see `.env.example`).

## Architecture

This is a PostgreSQL-backed Go service using goose for migrations and sqlc for type-safe query generation.

**Schema source of truth**: `db/migrations/` — goose migration files are what sqlc parses to derive the current schema. Make schema changes by adding/editing migrations; there is no separate `db/schema/` source of truth in this repository, and any documentation that mentions `db/schema/` should be treated as outdated.

**Query authoring flow**: write a named SQL query in `db/query/`, run `make sqlc-generate`, then use the generated Go functions from `internal/db/`. The generated package is named `db` with `pgx/v5` as the SQL driver and includes a `Querier` interface (`emit_interface: true`) suitable for mocking in tests.

**Entry point**: `cmd/server/main.go`
