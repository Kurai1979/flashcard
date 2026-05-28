---
name: best-practices-review
description: Review work on this Go service for best practices and security issues. Use when the user asks to "review my work", "check for issues", "look for best practices / security problems", or before opening a PR on this repo.
---

# Best-practices & security review

Review the current changes on this Go + chi + pgx + sqlc service for correctness, Go best practices, and security issues. Report findings grouped by severity (High / Medium / Low); do **not** make edits unless the user explicitly asks.

## Scope the diff first

```bash
go build ./... && go vet ./...          # vet catches real bugs (e.g. slog key/value mismatches)
git diff main...HEAD --stat             # what changed on this branch
```

Always run `go build` and `go vet` and report their output — `go vet` findings are real bugs, not style nits.

## What to check

**HTTP handlers (`internal/handlers/`)**
- Input validation: required fields, email format, password strength (min length).
- Email normalization (lowercase/trim) — Postgres `TEXT UNIQUE` is case-sensitive, so unnormalized emails allow near-duplicate accounts.
- Request body limits: `http.MaxBytesReader` + `json.Decoder.DisallowUnknownFields()`.
- Error responses must not leak internals (stack traces, driver errors, SQL). Use the `serverError` helper for 500s.
- Correct status codes: 201 on create, 400 on bad input, 409 on duplicate, 503 when a healthcheck dependency is down.
- Check-then-insert (TOCTOU) races: a duplicate-email pre-check can't replace handling the `UNIQUE` violation. Map pg error code `23505` to 409.
- Health/readiness endpoints must actually fail (non-2xx) when the DB is unreachable — not log-and-return-200.

**Auth (`internal/auth/`)**
- Argon2id params are sane (memory/iterations/parallelism), salt is random per-hash, comparison is constant-time (`subtle.ConstantTimeCompare`).
- Passwords/hashes are never logged.

**Data layer (`internal/db/`, `db/query/`)**
- Generated code is in sync — if `db/query/*.sql` changed, confirm `make sqlc-generate` was run.
- INSERTs that explicitly list columns with DB defaults (e.g. `is_active`) override those defaults with Go zero values. Either drop the column from the query so the default applies, or set the field explicitly in the handler.
- Queries are parameterized (sqlc enforces this — flag any raw string-built SQL).

**Server wiring (`cmd/server/main.go`)**
- `pgxpool.New` is lazy: add `pool.Ping(ctx)` at startup to fail fast, `defer pool.Close()`, and `os.Exit(1)` on fatal init errors instead of logging and continuing.
- `slog` takes key/value pairs, not printf args (`vet` flags mismatches).
- Config parsing (e.g. `PORT`) must return a safe default on error rather than falling through with a zero value.
- Set server timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`) rather than using the bare `http.ListenAndServe` defaults; consider graceful shutdown on SIGINT/SIGTERM.

**General Go hygiene**
- Errors wrapped with `%w` and context; no ignored errors on writes that matter.
- Context propagated from the request (`r.Context()`).
- No secrets in logs or committed to `.env` (only `.env.example`).

## Output

For each finding give: severity, `file:line`, what's wrong, and the concrete fix. End by listing which (if any) the user wants applied.
