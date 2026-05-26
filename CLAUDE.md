# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Tech stack

- **Language**: Go
- **HTTP router**: chi (`github.com/go-chi/chi/v5`)
- **Database**: PostgreSQL via `pgx/v5`
- **Migrations**: goose — `db/migrations/`
- **SQL → Go**: sqlc — type-safe query code generated into `internal/db/`
- **HTML templates**: templ — `.templ` files compiled to Go
- **Frontend interactivity**: htmx — server returns HTML fragments; no SPA/JS build step

## Tools

Install CLI tools before developing:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/a-h/templ/cmd/templ@latest
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

# templ (frontend templates)
templ generate       # compile *.templ -> *_templ.go (run after editing any .templ)
templ generate --watch   # rebuild on save during development
```

`DATABASE_URL` is read from `.env` (see `.env.example`).

## Architecture

This is a PostgreSQL-backed Go service using goose for migrations and sqlc for type-safe query generation.

**Schema source of truth**: `db/migrations/` — goose migration files are what sqlc parses to derive the current schema. Make schema changes by adding/editing migrations; there is no separate `db/schema/` source of truth in this repository, and any documentation that mentions `db/schema/` should be treated as outdated.

**Query authoring flow**: write a named SQL query in `db/query/`, run `make sqlc-generate`, then use the generated Go functions from `internal/db/`. The generated package is named `db` with `pgx/v5` as the SQL driver and includes a `Querier` interface (`emit_interface: true`) suitable for mocking in tests.

**HTTP layer**: chi router wired in `cmd/server/main.go`. Request handlers live in `internal/handlers/`; the `Handler` struct holds a `db.Querier` and a `*slog.Logger`.

**Frontend**: server-rendered HTML with templ + htmx — there is no JavaScript bundler or SPA. templ `.templ` files are the source of truth for markup; running `templ generate` compiles each into a `*_templ.go` file (committed alongside the source) exposing a `templ.Component`. Handlers render a component by calling its `.Render(ctx, w)`. htmx attributes (`hx-get`, `hx-post`, `hx-target`, …) in the templates drive partial-page updates; handlers respond to htmx requests with HTML fragments rather than JSON. The compiled `*_templ.go` files are generated — never edit them by hand; edit the `.templ` and regenerate.

**Entry point**: `cmd/server/main.go`
