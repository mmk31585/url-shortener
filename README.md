# URL Shortener

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat&logo=docker)

A production-oriented URL shortening REST API written in Go using only the standard library for HTTP serving. It creates short, trackable links backed by PostgreSQL, with redirect statistics, structured logging, request tracing, and an interactive Swagger UI.

```
POST http://localhost:8080/api/v1/urls   {"url": "https://example.com/very/long/path"}
  → 201  { "short_code": "Ab3xYk9Q", "short_url": "http://localhost:8080/Ab3xYk9Q", ... }

GET http://localhost:8080/Ab3xYk9Q        → 302 Location: https://example.com/very/long/path
```

## Features

- **Shorten URLs** — auto-generated, collision-safe 8-character base62 shortcodes (`crypto/rand`)
- **Redirect** — `302 Found` with a `Location` header and an atomically incremented click counter
- **Full CRUD** — create, read, update, soft-delete, and statistics per shortcode
- **Soft delete** — deleted links return `410 Gone`; a partial unique index allows shortcode reuse
- **Validation** — `net/url`-based checks (scheme, host, max length) at the API and database layers
- **Structured logging** — `log/slog` with JSON output in production, text in development
- **Request tracing** — per-request UUID via middleware, correlated across logs and error responses
- **Observable** — `/health` endpoint for liveness/readiness probes
- **Documented** — interactive Swagger UI served at `/swagger/`
- **Containerized** — multi-stage Docker build and a `docker-compose` stack for local development

## Tech Stack

| Concern | Choice | Why |
|---|---|---|
| Language | Go 1.26 | Single static binary, strong stdlib, `log/slog`, method-based routing |
| HTTP | `net/http` + `http.ServeMux` | Zero-dependency, production-ready; Go 1.22+ method/path patterns |
| Database | PostgreSQL 17 | Relational integrity, partial indexes, mature Go driver |
| Driver | `jackc/pgx/v5` (`database/sql`) | Parameterized queries, connection pooling |
| Migrations | `pressly/goose` | Numbered, reversible SQL migration files |
| Validation | `go-playground/validator` + `internal/validator` | Struct tags for payloads, pure functions for URLs/shortcodes |
| Logging | `log/slog` | Structured logging from the standard library |
| API docs | `swaggo/http-swagger` + swag | Annotation-driven Swagger UI |

## Getting Started

### Prerequisites

- Docker + Docker Compose (easiest path), or
- Go 1.26+ and a running PostgreSQL 17 instance

### Quick start with Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

This starts the API (`http://localhost:8080`) and a PostgreSQL 17 container (host port `5430`). The API waits for the database health check before serving.

Verify it works:

```bash
# Health check
curl http://localhost:8080/health

# Create a short URL
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/very/long/path?q=1"}'

# Follow the redirect (replace the shortcode with the one returned above)
curl -i http://localhost:8080/Ab3xYk9Q

# Interactive API docs
open http://localhost:8080/swagger/
```

### Run the API locally (no Docker for the app)

```bash
docker compose up -d db        # PostgreSQL only, exposed on localhost:5430
cp .env.example .env
make run
```

### Configuration

Configuration is loaded from environment variables (with an optional `.env` file). See `.env.example` for the full set.

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development`, `production`, or `test`; controls default log format |
| `SERVER_ADDRESS` | `:8080` | TCP address to listen on |
| `DB_ADDR` | *required* | PostgreSQL DSN, e.g. `postgres://user:pass@localhost:5430/db?sslmode=disable` |
| `DB_MAX_OPEN_CONNS` | `30` | Max open connections in the pool |
| `DB_MAX_IDLE_CONNS` | `30` | Max idle connections in the pool |
| `DB_MAX_LIFE_TIME` | `5m` | Max idle time per connection (`time.ParseDuration` format) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | auto | `text` or `json`; defaults to `json` when `APP_ENV=production` |
| `SHORTCODE_LENGTH` | `8` | Shortcode length; must be `8` |
| `MAX_URL_LENGTH` | `2048` | Maximum accepted URL length |

`POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD` configure the Docker Compose database container only.

## API

Base URL: `http://localhost:8080/api/v1` — Redirect path: `/{shortcode}`

| Method | Path | Description | Success |
|---|---|---|---|
| `POST` | `/api/v1/urls` | Create a short URL | `201 Created` |
| `GET` | `/api/v1/urls/{shortcode}` | Get URL info | `200 OK` |
| `PUT` | `/api/v1/urls/{shortcode}` | Change the destination URL | `200 OK` |
| `DELETE` | `/api/v1/urls/{shortcode}` | Soft-delete a short URL | `204 No Content` |
| `GET` | `/api/v1/urls/{shortcode}/stats` | Get redirect statistics | `200 OK` |
| `GET` | `/{shortcode}` | Redirect to the original URL | `302 Found` |
| `GET` | `/health` | Service and database health | `200 OK` |

### Create a short URL

```http
POST /api/v1/urls
Content-Type: application/json

{"url": "https://example.com/very/long/path?q=1"}
```

```http
HTTP/1.1 201 Created

{
  "id": 1,
  "short_code": "Ab3xYk9Q",
  "short_url": "http://localhost:8080/Ab3xYk9Q",
  "original_url": "https://example.com/very/long/path?q=1",
  "redirect_count": 0,
  "created_at": "2026-08-06T10:30:00Z",
  "updated_at": "2026-08-06T10:30:00Z"
}
```

### Redirect

```http
GET /Ab3xYk9Q

HTTP/1.1 302 Found
Location: https://example.com/very/long/path?q=1
Cache-Control: no-store
```

`302 Found` (not `301`) is used deliberately: browsers do not cache it, so the redirect counter stays accurate and destinations can change later. The `Cache-Control: no-store` header prevents caching of the redirect.

### Get / Update / Delete / Stats

```http
GET /api/v1/urls/Ab3xYk9Q
GET /api/v1/urls/Ab3xYk9Q/stats
PUT /api/v1/urls/Ab3xYk9Q        {"url": "https://new-destination.com"}
DELETE /api/v1/urls/Ab3xYk9Q
```

Delete is a soft delete: the row stays in the database with `deleted_at` set, the shortcode becomes reusable, and the redirect path returns `410 Gone`.

### Health check

```http
GET /health

HTTP/1.1 200 OK
{"status":"ok","database":"connected","timestamp":"2026-08-06T10:30:00Z"}
```

Returns `503 SERVICE_UNAVAILABLE` when the database cannot be reached.

### Errors

All errors share a consistent envelope:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "url not found",
    "request_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
  }
}
```

| Code | HTTP Status | When |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Malformed request or invalid input |
| `NOT_FOUND` | 404 | Shortcode does not exist |
| `CONFLICT` | 409 | Resource conflict |
| `GONE` | 410 | Redirect target was soft-deleted |
| `UNPROCESSABLE_ENTITY` | 422 | Payload fails validation |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
| `SERVICE_UNAVAILABLE` | 503 | Dependency (e.g. database) unavailable |

The `request_id` is generated by the request-ID middleware and appears in both the response and the application logs for correlation.

## Architecture

Layered architecture with strict one-way dependencies — business logic never touches HTTP or SQL directly.

```
cmd/api            entry point (composition root)
  ↓
handler            HTTP parsing, response formatting
  ↓
service            business logic, validation, error wrapping
  ↓
repository         URLRepository interface (mocks / PostgreSQL impl)
  ↓
domain             URL entity, ShortCode type, sentinel errors (no deps)
```

Supporting packages: `middleware` (request ID, logging, recovery), `shortener` (base62 generator strategy), `validator` (URL/shortcode rules), `logger` (`slog` setup), `config` (env loading), `storage`/`db` (pool + auto-migrations).

### Database

```sql
CREATE TABLE urls (
    id             SERIAL PRIMARY KEY,
    short_code     VARCHAR(255) NOT NULL,
    original_url   TEXT NOT NULL,
    redirect_count BIGINT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);
```

- **Partial unique index** — `UNIQUE INDEX ... ON urls(short_code) WHERE deleted_at IS NULL` enforces one active URL per shortcode while allowing soft-deleted shortcodes to be reused.
- **Check constraints** — non-empty URL, non-negative redirect count, exactly 8-character shortcode.
- **Migrations** — applied automatically on startup by Goose from `migrations/`.

## Project Structure

```
├── cmd/api/                 entry point: wiring, server, graceful shutdown
├── internal/
│   ├── config/              env-based configuration loading
│   ├── db/                  pgx connection pool (database/sql)
│   ├── domain/              URL entity, ShortCode, sentinel errors
│   ├── error/               AppError types and constructors
│   ├── handler/             HTTP handlers, DTOs, error writer
│   ├── integration/         end-to-end tests (build tag e2e)
│   ├── json/                JSON read/validate/write helpers
│   ├── logger/              slog setup (text/json, levels)
│   ├── middleware/          request ID, logging, recovery
│   ├── repository/          URLRepository interface
│   │   ├── mock/            in-memory implementation for tests
│   │   └── postgres/        PostgreSQL implementation
│   ├── router/              ServeMux routes + Swagger UI
│   ├── service/             URLService interface and implementation
│   ├── shortener/           Shortener strategy + RandomShortener
│   ├── storage/             DB connection + Goose auto-migrations
│   └── validator/           pure URL/shortcode validation
├── migrations/              Goose SQL migrations (up/down)
├── docs/                    generated Swagger spec (goes with /swagger/)
├── .github/workflows/       CI (audit) and release automation
├── Makefile                 common development commands
├── Dockerfile               multi-stage static binary build
└── docker-compose.yml       api + PostgreSQL stack
```

## Testing

```bash
make test          # unit + integration tests with the race detector
make test-e2e      # end-to-end HTTP tests (tags=e2e)
make coverage      # coverage report (opens coverage.html)
```

- Unit tests cover the validator, shortener, service, handlers, middleware, config, logger, and error packages using a hand-written mock repository.
- PostgreSQL repository tests run against a real database and skip automatically when `CI`/`GITHUB_ACTIONS` is unset and no database is reachable.
- E2E tests (`internal/integration`) exercise the full HTTP request/response cycle.

## Development

```bash
make build         # build ./cmd/api → bin/url-shortener
make run           # run the API from source
make fmt           # gofmt -s -w .
make vet           # go vet ./...
make lint          # go vet ./...
make swagger       # regenerate docs/ from swagger annotations
make tidy          # go mod tidy
make check         # swagger + tests + vet
make clean         # remove build artifacts
```

## CI / CD

- **Audit** (`.github/workflows/audit.yaml`) — runs on push/PR to `main`/`develop`: `go mod verify`, build, `go vet`, `staticcheck`, and `go test -race ./...`.
- **Release** (`.github/workflows/release-please.yaml`) — [release-please](https://github.com/googleapis/release-please) tags releases from Conventional Commit messages on `main`.

Use [Conventional Commits](https://www.conventionalcommits.org/) (e.g. `feat:`, `fix:`, `docs:`) so release notes are generated automatically.

## Design Notes

- **302 instead of 301** — keeps click counts accurate and destinations changeable.
- **Soft delete over hard delete** — audit trail and shortcode reuse via a partial unique index.
- **`crypto/rand` over `math/rand`** — non-guessable shortcodes.
- **Parameterized SQL everywhere** — no string interpolation, no SQL injection surface.
- **Full URLs are never logged** — query strings may contain tokens or PII; only metadata and error codes are logged.
- **Stateless service** — all state lives in PostgreSQL, so the API scales horizontally behind a load balancer without sticky sessions.


## roadmap idea project 

https://roadmap.sh/projects/url-shortening-service