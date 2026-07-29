# URL Shortening Service

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Architecture](https://img.shields.io/badge/Architecture-Layered%20%7C%20Clean%20-blue?style=flat)
![Database](https://img.shields.io/badge/Database-PostgreSQL-336791?style=flat&logo=postgresql)
![Testing](https://img.shields.io/badge/Coverage-80%25-green?style=flat)
![Docker](https://img.shields.io/badge/Container-Docker-2496ED?style=flat&logo=docker)

A production-oriented URL shortening REST API built with Go's standard library. Demonstrates Clean Architecture, SOLID principles, Domain-Driven Design, and layered backend engineering — from HTTP handling to PostgreSQL persistence, all without external frameworks.

---

## Table of Contents

1. [Project Vision](#project-vision)
2. [Software Requirements](#software-requirements)
3. [Domain Analysis](#domain-analysis)
4. [Architecture](#architecture)
5. [Technology Decisions](#technology-decisions)
6. [Project Structure](#project-structure)
7. [Database Design](#database-design)
8. [API Design](#api-design)
9. [Validation Rules](#validation-rules)
10. [Error Handling](#error-handling)
11. [Logging Strategy](#logging-strategy)
12. [Design Patterns](#design-patterns)
13. [SOLID Principles](#soli-principles)
14. [Clean Code](#clean-code)
15. [Security](#security)
16. [Performance](#performance)
17. [Scalability](#scalability)
18. [Testing Strategy](#testing-strategy)
19. [Development Roadmap](#development-roadmap)
20. [Git Strategy](#git-strategy)
21. [ADR — Architecture Decision Records](#adr--architecture-decision-records)
22. [Deployment](#deployment)
23. [Future Improvements](#future-improvements)
24. [Learning Outcomes](#learning-outcomes)
25. [Engineering Notes](#engineering-notes)
26. [Final Project Checklist](#final-project-checklist)

---

## Project Vision

### Problem Statement

Long URLs are impractical for sharing in social media posts, email campaigns, SMS messages, QR codes, and printed materials. A URL shortening service solves this by transforming opaque long URLs into compact, trackable short links (e.g., `https://short.io/abc123`). Beyond brevity, the service provides access statistics, brand control, and centralized link management.

### Business Value

- **Character savings**: Every character saved in a social media post is a character available for message content.
- **Branding**: Custom short domains (`nyti.ms`, `t.co`) reinforce brand identity.
- **Analytics**: Click counts, referrer data, and geolocation enable marketing measurement.
- **Flexibility**: Changing a link's destination without redistributing the short URL.
- **QR utility**: Short URLs fit into QR codes with room for additional data.

### Goals

1. Build a fully functional URL shortening REST API in Go using only the standard library
2. Demonstrate SOLID principles, Clean Code, and layered architecture in a real project
3. Provide production-grade patterns (repository pattern, dependency injection, structured logging)
4. Include comprehensive testing (unit + integration at 80%+ coverage)
5. Containerize with Docker for reproducible deployments
6. Serve as a learning journal and engineering reference that stands up to recruiter scrutiny

### Scope

**In Scope:**

- Create short URLs (auto-generated 8-character base62 shortcodes)
- Retrieve original URL and metadata by shortcode
- Redirect to original URL (HTTP 302)
- Update existing URL destination
- Soft-delete short URLs (with `deleted_at`)
- Access statistics (redirect count)
- Input validation (URL format, length)
- Structured logging with `log/slog`
- Health check endpoint
- Swagger/OpenAPI documentation
- PostgreSQL persistence with Goose migrations
- Docker support (multi-stage build)
- Unit and integration testing

**Out of Scope:**

- Authentication / Authorization
- Custom aliases (user-chosen shortcodes)
- URL expiration / TTL
- Rate limiting
- Redis caching
- QR code generation
- Analytics dashboard (frontend)
- Multi-tenant support
- Message queues or background workers
- CDN integration
- Microservices

### Target Users

- **Developers**: Integrate the API into their applications
- **Marketers**: Create trackable campaign links
- **Content creators**: Share clean, professional links
- **QA engineers**: Test the service and validate behavior

### Real-World Examples

| Service | Owner | Scale |
|---|---|---|
| Bitly | Bitly Inc. | 600M+ links |
| TinyURL | TinyURL Inc. | 200M+ links |
| Rebrandly | Rebrandly | Enterprise-focused |
| Firebase Dynamic Links | Google | Platform-integrated |
| t.co | Twitter/X | All tweets processed |

### Success Criteria

1. `go build ./cmd/server` produces a working binary
2. All 6 CRUD endpoints respond correctly with proper HTTP status codes
3. Redirect endpoint returns 302 with correct Location header
4. Redirect count increments atomically on every successful redirect
5. Soft-deleted URLs return 410 Gone on redirect attempt
6. 80%+ test coverage across all packages
7. Docker image builds and runs successfully with `docker-compose up`
8. Swagger UI is accessible and documents all endpoints
9. Structured JSON logs are emitted to stdout in production mode
10. All SOLID principles are demonstrably applied in the codebase

### Checklist

- [ ] Problem statement documented with business value
- [ ] Goals defined (6 specific, measurable goals)
- [ ] Scope explicitly listed (in/out)
- [ ] Target users identified
- [ ] Real-world competitor examples catalogued
- [ ] Success criteria defined (10 measurable criteria)

### References

- See [Software Requirements](#software-requirements) for formal requirements
- See [Development Roadmap](#development-roadmap) for implementation plan

---

## Software Requirements

### Functional Requirements

| ID | Requirement | Description | Priority |
|---|---|---|---|
| FR-01 | Create Short URL | Accept a long URL, validate it, generate a unique 8-char base62 shortcode, store the mapping, and return the short URL | Must |
| FR-02 | Redirect | When a client requests `/r/{shortcode}`, look up the original URL and respond with HTTP 302 Found | Must |
| FR-03 | Get URL Info | Retrieve the details of a shortened URL by its shortcode | Must |
| FR-04 | List URLs | Return all active (non-deleted) short URLs | Should |
| FR-05 | Update URL | Change the original URL for an existing shortcode | Must |
| FR-06 | Delete URL | Soft-delete a short URL by setting `deleted_at` | Must |
| FR-07 | Statistics | Return access statistics (redirect count, creation metadata) | Should |
| FR-08 | URL Validation | Validate that submitted URLs have a valid scheme and host | Must |
| FR-09 | Health Check | Return service health including database connectivity | Must |
| FR-10 | Swagger | Document all API endpoints with Swagger/OpenAPI | Should |
| FR-11 | Structured Logging | Emit structured JSON logs to stdout | Must |
| FR-12 | Concurrency Safety | Handle concurrent requests safely via database-level locking | Must |

### Non-Functional Requirements

| ID | Category | Requirement | Metric |
|---|---|---|---|
| NFR-01 | Performance | Redirect latency | < 50ms p99 |
| NFR-02 | Performance | Create URL latency | < 200ms p99 |
| NFR-03 | Scalability | Handle concurrent requests | 10,000+ concurrent |
| NFR-04 | Reliability | Data durability | PostgreSQL WAL + replication |
| NFR-05 | Maintainability | Code coverage | >= 80% |
| NFR-06 | Maintainability | Zero external dependencies except pgx and goose | Must |
| NFR-07 | Extensibility | Repository interface swap-able | Any storage backend |
| NFR-08 | Security | No SQL injection | Parameterized queries only |
| NFR-09 | Observability | Structured logging | slog JSON in production |
| NFR-10 | Deployment | Docker build | < 50MB final image |
| NFR-11 | Availability | Uptime target | 99.9% |
| NFR-12 | Portability | Static binary | CGO_ENABLED=0 |

### Business Requirements

| ID | Rule | Description | Enforcement |
|---|---|---|---|
| BR-01 | ShortCode uniqueness | Each shortcode maps to exactly one URL at a time | DB UNIQUE constraint |
| BR-02 | URL validity | Original URL must have http/https scheme and non-empty host | Application + DB CHECK constraint |
| BR-03 | Auto-generation | Shortcodes are generated automatically using crypto/rand + base62 | Application layer |
| BR-04 | Redirect counting | Counter increments only after successful redirect delivery | Atomic UPDATE in repository |
| BR-05 | Shortcode reuse | Deleted shortcodes may be reused for new URLs | Application allows re-creation of deleted shortcodes |
| BR-06 | Soft delete | Deleted URLs remain in DB with `deleted_at` set | Application layer + queries filter `deleted_at IS NULL` |
| BR-07 | Fixed shortcode length | All shortcodes are 8 characters | Repository shortener strategy |
| BR-08 | Immutable created_at | `created_at` is set once and never updated | Application layer |
| BR-09 | Updated timestamp | `updated_at` changes on every write operation | Application layer |
| BR-10 | No custom aliases | Users cannot choose their own shortcode in this version | Application rejects alias attempts |

### System Constraints

- **Language**: Go 1.24+
- **HTTP**: `net/http` standard library only (no Gin, Echo, Fiber, Chi)
- **Database**: PostgreSQL 16+
- **Migration**: Goose v3
- **Logging**: `log/slog` standard library only (no zap, zerolog, logrus)
- **Driver**: `github.com/jackc/pgx/v5` for PostgreSQL, `github.com/pressly/goose/v3` for migrations
- **No ORM**: All SQL is hand-written, parameterized, and explicit
- **No framework**: Pure standard library; any external dependencies must be justified

### Assumptions

1. The database (PostgreSQL) is available and accessible at configuration time
2. The Go toolchain is installed and the project is built with Go 1.24+
3. Shortcodes are opaque to clients — they contain no encoded information
4. The service will run as a single process behind a load balancer (stateless)
5. Redirects are temporary (302), not permanent (301) — this allows future destination changes
6. The `original_url` field in the database is not encrypted — this is acceptable for an MVP
7. The service handles 10K concurrent requests is sufficient for the target scale

### Acceptance Criteria

- [ ] POST `/api/v1/urls` with valid URL creates a short URL and returns 201
- [ ] GET `/r/{shortcode}` returns HTTP 302 with correct Location header
- [ ] GET `/api/v1/urls/{shortcode}` returns URL details with 200
- [ ] PUT `/api/v1/urls/{shortcode}` updates destination and returns 200
- [ ] DELETE `/api/v1/urls/{shortcode}` soft-deletes and returns 204
- [ ] GET `/api/v1/urls/{shortcode}/stats` returns redirect count
- [ ] GET `/health` returns 200 with database connection status
- [ ] Invalid URL submissions return 400 with descriptive error
- [ ] Nonexistent shortcodes return 404
- [ ] Soft-deleted shortcodes return 410 on redirect
- [ ] All error responses follow consistent format
- [ ] Structured JSON logs are emitted to stdout
- [ ] All tests pass with `go test -race -cover ./...`
- [ ] Docker build produces a small static binary
- [ ] Swagger UI is accessible and documents all endpoints

### Checklist

- [ ] All 12 functional requirements defined with IDs
- [ ] All 12 non-functional requirements defined with measurable metrics
- [ ] All 10 business rules documented with enforcement mechanism
- [ ] All 7 system constraints listed
- [ ] All 7 assumptions documented
- [ ] All 13 acceptance criteria defined
- [ ] Out of scope explicitly listed

### References

- See [Domain Analysis](#domain-analysis) for business rule implementation in entities
- See [API Design](#api-design) for how requirements map to HTTP endpoints
- See [Development Roadmap](#development-roadmap) for which phase implements each requirement
- See [ADR](#adr--architecture-decision-records) for technology decision rationale

---

## Domain Analysis

### Domain Model

The domain model represents the core business entities in their purest form — without any knowledge of HTTP, databases, or frameworks. These are plain Go structs that carry the business meaning of the domain.

### Entities

#### URL (Primary Entity)

The `URL` entity is the central domain object. It represents the mapping between a shortcode and its original destination.

| Field | Type | Description |
|---|---|---|
| ID | `int64` | Internal database identifier (SERIAL) |
| ShortCode | `string` (8 chars) | Public-facing base62 token |
| OriginalURL | `string` | The full destination URL |
| RedirectCount | `int64` | Number of successful redirects |
| CreatedAt | `time.Time` | When the short URL was created |
| UpdatedAt | `time.Time` | When the record was last modified |
| DeletedAt | `*time.Time` (nullable) | Nil if active, set on soft delete |

**State transitions:**

```
[Active] --(redirect)--> [Active]  (count increments)
[Active] --(update)-->   [Active]  (destination changes, UpdatedAt changes)
[Active] --(delete)-->   [Deleted] (DeletedAt set)
[Deleted] --(recreate)-->[Active]
```

**Invariants:**

- `ShortCode` is always exactly 8 characters
- `ShortCode` is unique across all active URLs (soft-deleted shortcodes can be reused)
- `OriginalURL` must have `http://` or `https://` scheme and a non-empty host
- `RedirectCount` is always >= 0
- `CreatedAt` is set once at creation and never changes
- `UpdatedAt` is set on every write operation (create, update, soft-delete)
- `DeletedAt` is either nil (active) or a past timestamp (deleted)

### Use Cases (Primary Actors: Client)

```mermaid
graph TD
    Client["Client"] --> UC1["Create Short URL"]
    Client --> UC2["Redirect via Short URL"]
    Client --> UC3["Get URL Info"]
    Client --> UC4["Update URL Destination"]
    Client --> UC5["Delete URL"]
    Client --> UC6["Get Statistics"]
    Client --> UC7["Health Check"]
    Client --> UC8["List All URLs"]

    UC1 -->|"generates"| ShortCode["8-char base62 shortcode"]
    UC2 -->|"increments"| Count["redirect count + 1"]
    UC4 -->|"updates"| Destination["original URL"]
    UC5 -->|"sets"| DeletedAt["deleted_at = NOW()"]
```

### Mermaid ER Diagram

```mermaid
erDiagram
    URL {
        bigint id PK "auto-increment, internal only"
        varchar short_code UK "unique, 8 chars base62"
        text original_url NOT NULL "http/https URL"
        bigint redirect_count "counter, >=0"
        timestamp created_at "UTC, immutable"
        timestamp updated_at "UTC, changes on write"
        timestamp deleted_at "nullable, deletion marker"
    }
```

### Business Rules (Detailed)

**BR-01 — ShortCode Uniqueness**
The `short_code` column has a `UNIQUE` constraint at the database level. Auto-generation using crypto/rand with 8 base62 characters produces 62^8 ≈ 218 trillion possible combinations. The probability of collision is negligible for the target scale, but the database constraint serves as the final safety net against race conditions.

**BR-02 — URL Validity**
The `net/url` package is used for parsing (not regex). A URL must have scheme `http` or `https` and a non-empty host. Maximum length is 2048 characters (browser URL bar limit). This is validated at the application layer before any database operation.

**BR-03 — Auto-Generation**
The `Shortener` strategy interface abstracts code generation. The default `RandomShortener` uses `crypto/rand.Read()` to fill a byte slice, then maps each byte to the base62 alphabet. This is a deterministic process per random input, and the uniqueness guarantee comes from the database UNIQUE constraint.

**BR-04 — Redirect Counting**
The redirect count increments atomically using `UPDATE urls SET redirect_count = redirect_count + 1, updated_at = NOW() WHERE short_code = $1 AND deleted_at IS NULL RETURNING *`. This ensures the count is accurate even under concurrent access. The increment happens in the same database operation as the SELECT for the redirect, avoiding a two-step race.

**BR-05 — Shortcode Reuse**
When a URL is soft-deleted, the `deleted_at` column is set. Future GET operations filter by `deleted_at IS NULL`, making the shortcode available for a new URL with the same shortcode value. The UNIQUE constraint is enforced with a partial index: `CREATE UNIQUE INDEX idx_urls_short_code ON urls(short_code) WHERE deleted_at IS NULL` — this allows the same shortcode to exist for active and deleted URLs simultaneously, but only one active URL per shortcode.

**BR-06 — Soft Delete**
Deletion sets `deleted_at = NOW()`. No rows are physically removed. All queries for active URLs include `WHERE deleted_at IS NULL`. This provides a safety net against permanent data loss and enables analytics on deleted URLs.

**BR-07 — Fixed Shortcode Length**
All shortcodes are 8 characters using the base62 alphabet (a-z, A-Z, 0-9). Fixed length simplifies generation, validation, and URL routing.

**BR-08 — No Custom Aliases**
Custom alias support is deferred to a future version. This prevents namespace conflicts, validation complexity around reserved words, and the need for a namespace management system.

**BR-09 — Immutable created_at**
The `created_at` timestamp is set once on INSERT and never modified. The application does not update this field on any subsequent write.

**BR-10 — Updated Timestamp**
`updated_at` is set to the current time on every write (INSERT, UPDATE, soft DELETE). This provides an audit trail of when records were last modified.

### Engineering Notes

The domain model intentionally has no methods — it is a pure data structure. All behavior (validation, shortcode generation, redirect counting) lives in the Service layer. This keeps the domain model clean, testable in isolation, and decoupled from persistence concerns.

The partial unique index on `short_code WHERE deleted_at IS NULL` is a critical database-level decision. Without it, the `UNIQUE` constraint on `short_code` would prevent reuse of soft-deleted shortcodes entirely. With a partial index, deleted shortcodes become available for new allocations while maintaining uniqueness for active URLs.

### Common Mistakes

- **Putting business logic in the URL model**: The model is a data container. Never add methods like `Validate()`, `IncrementCount()`, or `IsDeleted()` to the model struct. These belong in the Service layer.
- **Using a boolean `is_deleted` instead of `deleted_at`**: A timestamp provides audit trail capability and enables future "restore deleted URL" features.
- **Not filtering `deleted_at IS NULL` in read queries**: Soft-deleted URLs will appear in search results and redirects, defeating the purpose of deletion.
- **Allowing the UNIQUE constraint to prevent shortcode reuse**: Without the partial index approach, attempting to create a new URL with a reuseable shortcode would fail with a unique constraint violation.
- **Incrementing redirect count in the service layer instead of atomically in SQL**: If the counter increment and the redirect retrieval are separate database calls, concurrent requests can cause lost updates or incorrect counts.

### Senior Tips

Store business rules in a single document and reference them by ID (BR-01, BR-02) in code comments and test names. This makes it easy to trace a test back to the rule it validates. Consider adding a `BUSINESS_RULES.md` file that maps each rule ID to its implementation location in the codebase.

### Checklist

- [ ] URL entity defined with all 7 fields and correct types
- [ ] State transitions documented (Active → Deleted → Active)
- [ ] All 10 business rules (BR-01 to BR-10) documented
- [ ] Mermaid ER diagram created
- [ ] Use case diagram created (mermaid)
- [ ] Primary actor identified (Client)
- [ ] Each entity's invariants documented
- [ ] ShortCode uniqueness explained with database constraint rationale
- [ ] Soft delete pattern explained with partial index
- [ ] Redirect count atomicity explained

### References

- See [Database Design](#database-design) for schema and index implementation
- See [Validation Rules](#validation-rules) for application-layer enforcement
- See [Service Design](#development-roadmap) for where each rule is enforced
- See [ADR](#adr--architecture-decision-records) for database and pattern decisions

---

## Architecture

### Architecture Style: Layered Architecture

Layered architecture separates the application into horizontal tiers, each with a single responsibility. Components in each layer communicate only with the layer directly below them.

**Why Layered Architecture:**
For a URL shortener with 6 primary endpoints and one storage backend, layered architecture provides the optimal balance of separation of concerns and simplicity. It avoids the ceremony of Clean Architecture (Hexagonal) — ports, adapters, domain services, application services — which adds value in complex domains with multiple external systems but is over-engineered here.

**Why Not Clean Architecture / Hexagonal:**
Clean Architecture's additional layers (Domain, Use Cases, Ports/Adapters) add value when you have complex business rules, multiple external integrations (payment processing, email, search), or a team of 10+ developers. For a URL shortener with 6 endpoints and one database, these additional layers create unnecessary indirection that makes the code harder to read and trace without proportionally improving testability or flexibility.

**Why Not Monolithic (everything in main.go):**
A monolithic approach makes testing impossible (no layer can be isolated), prevents future extensibility (adding a new feature requires modifying the entire file), and violates Single Responsibility Principle.

### Architecture Decisions

| Decision | Reasoning |
|---|---|
| Layered Architecture | Separates concerns into 4 tiers, optimal for this scope |
| Repository Pattern | Abstracts data access, enables testing with mocks, allows storage backend swaps |
| Dependency Injection | Explicit wiring in main.go, no service locator or DI framework needed |
| Strategy Pattern | Shortcode generation algorithm is isolated and swappable |
| net/http standard library | Zero dependencies, demonstrates Go idioms, production-ready |
| log/slog standard library | Structured logging with zero external deps, JSON/text handlers |
| PostgreSQL | Relational integrity, JSON support, mature Go driver |
| Goose migrations | Numbered SQL files, up/down support, mature and widely used |
| Hand-written SQL | No ORM magic, explicit, transparent, testable |

### Layer Responsibilities

**Presentation Layer (`handler/`)**
- Receives HTTP requests and parses input (JSON body, URL path parameters)
- Calls the appropriate service method
- Formats HTTP responses (status codes, JSON body, headers)
- Knows about HTTP but nothing about databases or business rules
- Thin layer — delegates all logic to service

**Service Layer (`service/`)**
- Contains all business logic
- Orchestrates validation, shortcode generation, and data access
- Manages transactions and error wrapping
- Has zero imports from `net/http`, `database/sql`, or any external package
- Depends on the Repository interface only

**Repository Layer (`repository/`)**
- Defines the `URLRepository` interface (the abstraction)
- Concrete implementation in `repository/postgres/`
- Translates domain entities to SQL queries and database rows back to domain entities
- Maps database errors to application-level errors at the boundary
- Uses parameterized queries exclusively

**Storage Layer (`migrations/`, PostgreSQL)**
- Database schema, indexes, and constraints defined in SQL migration files
- Managed by Goose (numbered, ordered, reversible)
- Enforces data integrity at the database level (UNIQUE, CHECK, NOT NULL)

### Dependency Direction

```
cmd/server/ (entry point)
    ↓ depends on
handler/ (presentation)
    ↓ depends on
service/ (business logic)
    ↓ depends on
repository/ (abstraction + postgres/ implementation)
    ↓ depends on
domain/ (entities, no external deps)
    ↓ depends on
storage/ (PostgreSQL, migrations)
```

**The golden rule:** Dependencies always point inward (toward domain). Higher layers depend on lower layers via interfaces. Lower layers never import higher layers.

**The dependency graph in practice:**
```
cmd → handler → service → repository (interface) → domain
                                               ↑
                                    repository/postgres (concrete)
                                    ↓ depends on
                                    storage (database/sql)
```

The service imports only the `repository` interface package, never `repository/postgres`. The concrete PostgreSQL implementation imports the `repository` interface package and `domain`. This ensures that business logic is completely decoupled from storage technology.

### Data Flow (Full Request Lifecycle)

**Create URL (Write Path):**
```
1. Client: POST /api/v1/urls {"url": "https://example.com/..."}
2. Router matches POST /api/v1/urls → CreateURLHandler
3. Handler reads request body, parses JSON
4. Handler calls validation.ValidateURL(request.URL)
5. If validation fails → 400 response, handler returns
6. Handler calls service.CreateURL(ctx, originalURL)
7. Service validates URL again (defense in depth)
8. Service calls shortener.Generate() to get a unique shortcode
9. Service calls repository.Create(ctx, urlEntity)
10. Repository executes parameterized INSERT query
11. PostgreSQL inserts row, returns generated id
12. Repository scans row into domain.URL, returns
13. Service returns domain.URL to handler
14. Handler formats response JSON (includes short_url)
15. Handler writes 201 Created with JSON body
16. Middleware logs request: method, path, status, duration, request_id
```

**Redirect (Read Path):**
```
1. Client: GET /r/abc123de
2. Router matches GET /r/{shortcode} → RedirectHandler
3. Handler extracts shortcode from URL path
4. Handler calls service.GetByShortCode(ctx, "abc123de")
5. Service calls repository.GetByShortCode(ctx, "abc123de")
6. Repository executes: SELECT * FROM urls WHERE short_code = $1 AND deleted_at IS NULL
7. PostgreSQL returns row (or empty)
8. Repository scans row into domain.URL, returns
9. If not found → service returns ErrURLNotFound → handler returns 404
10. If found → service calls repository.IncrementRedirectCount(ctx, "abc123de")
11. Repository executes: UPDATE urls SET redirect_count = redirect_count + 1, updated_at = NOW() WHERE short_code = $1 AND deleted_at IS NULL RETURNING *
12. PostgreSQL updates row, returns updated URL
13. Service returns original URL to handler
14. Handler sets Location header, status 302 Found
15. Client follows redirect
16. Middleware logs request with duration and status 302
```

### Architecture Diagram

```mermaid
graph TD
    A["HTTP Client"] --> B["Handler Layer<br/>parsing, response formatting"]
    B --> C["Service Layer<br/>business logic, validation"]
    C --> D["Repository Interface<br/> abstraction"]
    D --> E["Postgres Repository<br/>concrete implementation"]
    E --> F["PostgreSQL<br/>persistence"]
    F -->|rows| E
    E -->|entity| D
    D -->|entity| C
    C -->|response DTO| B
    B -->|HTTP response| A

    subgraph "Presentation"
        B
    end
    subgraph "Business Logic"
        C
    end
    subgraph "Data Access"
        D
        E
    end
    subgraph "Persistence"
        F
    end
```

### Sequence Diagram (Create URL)

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler
    participant V as Validator
    participant S as Service
    participant SC as Shortener
    participant R as Repository
    participant DB as PostgreSQL

    C->>H: POST /api/v1/urls {url: "..."}
    H->>H: Parse JSON body
    H->>V: ValidateURL(url)
    alt Invalid URL
        V-->>H: ErrInvalidURL
        H-->>C: 400 Bad Request
    else Valid URL
        H->>S: CreateURL(ctx, url)
        S->>SC: Generate()
        SC-->>S: "abc123de"
        S->>R: Create(ctx, urlEntity)
        R->>DB: INSERT ... VALUES ($1,$2,...) RETURNING id
        DB-->>R: Row with id, timestamps
        R-->>S: URL entity
        S-->>H: URL entity
        H->>H: Format JSON response
        H-->>C: 201 Created {short_code, short_url, ...}
    end
```

### Sequence Diagram (Redirect)

```mermaid
sequenceDiagram
    participant C as Client
    participant H as RedirectHandler
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL

    C->>H: GET /r/abc123de
    H->>S: GetByShortCode(ctx, "abc123de")
    S->>R: GetByShortCode(ctx, "abc123de")
    R->>DB: SELECT * FROM urls WHERE short_code=$1 AND deleted_at IS NULL
    alt Not found
        DB-->>R: No rows
        R-->>S: ErrURLNotFound
        S-->>H: ErrURLNotFound
        H-->>C: 404 Not Found
    else Found (active)
        DB-->>R: URL row
        R-->>S: URL entity
        S->>R: IncrementRedirectCount(ctx, "abc123de")
        R->>DB: UPDATE urls SET redirect_count=redirect_count+1 WHERE short_code=$1 RETURNING *
        DB-->>R: Updated row
        R-->>S: Updated URL
        S-->>H: Original URL string
        H->>H: Set Location header, status 302
        H-->>C: 302 Found (Location: https://...)
    end
```

### Component Diagram

```mermaid
graph LR
    subgraph "cmd/server"
        main["main.go<br/>Composition root"]
    end

    subgraph "internal/handler"
        create["CreateURLHandler"]
        get["GetURLHandler"]
        redirect["RedirectHandler"]
        update["UpdateURLHandler"]
        delete["DeleteURLHandler"]
        stats["StatsHandler"]
        health["HealthHandler"]
    end

    subgraph "internal/service"
        svc["URLService<br/>business logic"]
    end

    subgraph "internal/repository"
        iface["URLRepository<br/>interface"]
        pg["PostgresURLRepository<br/>sql.DB, pgx"]
    end

    subgraph "internal/domain"
        url["URL entity"]
        errs["Domain errors"]
    end

    subgraph "internal/storage"
        goose["Goose migrations"]
        pgconn["PostgreSQL"]
    end

    subgraph "internal/shortener"
        gen["RandomShortener<br/>crypto/rand + base62"]
    end

    subgraph "internal/validator"
        val["ValidateURL<br/>net/url"]
    end

    subgraph "internal/middleware"
        log["Logging middleware"]
        rid["Request ID middleware"]
        rec["Recovery middleware"]
    end

    main --> create
    main --> redirect
    main --> svc
    create --> val
    create --> svc
    svc --> iface
    iface <-- implements --> pg
    pg --> pgconn
    pgconn -.-> goose
    svc --> gen
    svc --> url
    svc --> errs
    log -.-> redirect
```

### Engineering Notes

The sequence of dependency layers is the single most important architectural decision in this project. If the dependency graph is violated (e.g., service imports `database/sql` directly), the entire testability and flexibility benefit is lost. Always verify the dependency graph with `go mod graph` or by reading the imports of each package.

The two-step approach (SELECT then UPDATE) for redirects is simpler than a single UPSERT for the initial implementation. Under high concurrency, consider consolidating to `UPDATE ... RETURNING` which handles both the existence check and count increment atomically.

### Common Mistakes

- **Circular imports between service and repository**: If service imports the concrete postgres package instead of the interface, the dependency direction breaks. The service should only import the repository interface package.
- **Handler contains validation logic**: If your handler has `if strings.TrimSpace(url) == ""` or `if parsed.Scheme != "https"`, that validation should have been extracted to the validator package. The handler only checks if the request body is valid JSON.
- **Middleware ordering is wrong**: Middleware should be applied outer-to-inner. Recovery (innermost), then request ID, then logging (outermost). Logging middleware must be outermost to capture the response status code set by later middleware.

### Senior Tips

Start each new package by defining its interface before writing any implementation code. This forces you to think about the contract first and makes the dependency graph explicit from day one. Use `go list -f '{{.ImportPath}}' ./...` to verify which packages import which, and confirm no circular dependencies exist.

### Checklist

- [ ] Layered architecture selected and justified (vs Clean Architecture, monolithic)
- [ ] All 4 layers identified and responsibilities defined
- [ ] Dependency direction documented (Handler → Service → Repository → Storage)
- [ ] Mermaid architecture diagram created
- [ ] Mermaid sequence diagrams created (create URL + redirect)
- [ ] Mermaid component diagram created
- [ ] Full request lifecycle documented step by step for both read and write paths
- [ ] Why net/http over frameworks documented (ADR)
- [ ] Why slog over zap/zerolog documented (ADR)
- [ ] Why PostgreSQL over MySQL/SQLite/MongoDB documented (ADR)
- [ ] Why Goose over manual migrations documented (ADR)
- [ ] Why Repository Pattern over Active Record documented (ADR)
- [ ] Why Layered over Clean Architecture documented (ADR)
- [ ] Why SERIAL over UUID/BIGSERIAL documented (ADR)
- [ ] Data flow documented for write and read paths
- [ ] Context propagation documented through all layers

### References

- See [Technology Decisions](#technology-decisions) for ADRs
- See [Design Patterns](#design-patterns) for pattern implementations
- See [SOLID Principles](#soli-principles) for dependency inversion justification
- See [Project Structure](#project-structure) for where each layer lives

---

## Technology Decisions

Each technology selection is justified with alternatives considered and trade-offs explained.

### Go

| Aspect | Detail |
|---|---|
| **Why Go** | Compiled to a single static binary, excellent concurrency support (goroutines), strong standard library, fast compilation, growing ecosystem for backend services |
| **Why Go over Node.js** | Type safety catches bugs at compile time, no runtime dependency issues, better performance for CPU-bound tasks, simpler deployment (single binary) |
| **Why Go over Python** | No GIL (global interpreter lock), better performance, static typing, compiled binary with no runtime dependencies |
| **Why Go over Rust** | Faster development velocity, simpler memory model (garbage collected), adequate performance for this workload (I/O-bound HTTP service) |
| **Version** | Go 1.24+ for latest features including `log/slog` (Go 1.21+) and improved generic support |

### net/http (Standard Library)

| Aspect | Detail |
|---|---|
| **Why net/http** | Production-ready, zero dependencies, excellent performance, demonstrates Go idioms. The standard library HTTP server handles 10K+ concurrent connections effortlessly |
| **Why not Gin** | Gin adds an external dependency and teaches framework-specific patterns rather than Go idioms. For 6 endpoints, the additional complexity is not justified |
| **Why not Echo** | Similar trade-offs to Gin — framework-specific patterns, external dependency, unnecessary for this scope |
| **Why not Chi** | Chi is a lightweight router that works with net/http and could be adopted later for more complex routing needs. Not needed for 6 endpoints |
| **Trade-off** | `http.ServeMux` does not support path parameters (e.g., `/r/{shortcode}`). The project handles this by parsing path segments manually in handlers or by using a simple pattern-matching approach |

### log/slog (Standard Library Logger)

| Aspect | Detail |
|---|---|
| **Why slog** | Structured logging built into Go standard library since 1.21. JSON output for production, human-readable for development. Zero dependencies |
| **Why not zap** | zap is faster (zero-allocation JSON) but adds an external dependency. slog's performance is sufficient for this project's scale. zap can be adopted later if profiling shows a bottleneck |
| **Why not zerolog** | Similar to zap — excellent performance but external dependency. slog provides equivalent structured logging capabilities |
| **Why not logrus** | logrus is deprecated (no longer maintained) and lacks structured logging |
| **Trade-off** | slog is slightly slower than zap for high-throughput logging, but for the URL shortener's expected throughput (thousands of redirects/second), slog is more than adequate |

### PostgreSQL

| Aspect | Detail |
|---|---|
| **Why PostgreSQL** | Relational integrity (constraints, transactions), JSON support for future flexibility, mature Go driver (`pgx`), free/open-source, ACID compliance |
| **Why not MySQL** | PostgreSQL has better standards compliance, superior JSON support, and a more robust type system |
| **Why not SQLite** | Excellent for embedded use but lacks concurrent write support — not suitable for a production HTTP service with multiple concurrent writers |
| **Why not MongoDB** | Overkill for a single-entity schema, lacks relational guarantees for future features, adds operational complexity |
| **Trade-off** | PostgreSQL requires a separate database process. This adds operational overhead compared to embedded options but is necessary for production-grade reliability |

### Goose (Migrations)

| Aspect | Detail |
|---|---|
| **Why Goose** | Mature migration tool using numbered SQL files, supports up/down migrations for rollbacks, good PostgreSQL integration, simple file-based approach |
| **Why not golang-migrate** | Slightly more complex configuration, less Go community adoption for small projects |
| **Why not manual SQL scripts** | No version tracking, no rollback capability, error-prone for production deployments |
| **Why not ORM-generated migrations** | Hand-written SQL in Goose files are explicit, readable, reviewable, and make no assumptions about schema design |

### Repository Pattern

| Aspect | Detail |
|---|---|
| **Why Repository** | Abstracts data access behind an interface, enables mock repositories for testing, allows storage backend changes without modifying business logic |
| **Why not Active Record** | Active Record (models have save/update/delete methods) couples domain logic to persistence, violating Single Responsibility Principle |
| **Why not Data Mapper with ORM** | An ORM (like GORM) adds magic behavior that is hard to debug, implicit, and creates tight coupling. Hand-written SQL in the repository is explicit and transparent |
| **Why not no repository (service talks directly to db)** | This couples business logic to the database driver, making testing impossible without a real database |

### Dependency Injection

| Aspect | Detail |
|---|---|
| **Why Manual DI** | Dependencies are injected via constructor parameters in `main.go`. This is explicit, debuggable, and requires no framework |
| **Why not Service Locator** | Service Locator hides dependencies and makes them harder to trace, violating the explicit over implicit principle |
| **Why not DI Framework** | Frameworks like Uber's dig add unnecessary complexity for a project of this size. Manual DI forces you to think about each dependency explicitly |

### Strategy Pattern (Shortcode Generator)

| Aspect | Detail |
|---|---|
| **Why Strategy** | Shortcode generation algorithm is encapsulated behind an interface (`Shortener`), enabling different implementations (Random, Hash, Sequential) to be swapped without modifying the service |
| **Why not hardcoded function** | A function type `func() (string, error)` would work for a single algorithm but provides no extensibility if you want to add metadata, configuration, or stateful generation |
| **Trade-off** | For this simple project, Strategy may be over-engineered. However, it demonstrates the pattern and makes future extensibility (e.g., adding hash-based shortcodes for deduplication) trivial |

### Checklist

- [ ] Go selected with version rationale
- [ ] net/http selected over Gin/Echo/Chi with reasoning
- [ ] slog selected over zap/zerolog/logrus with reasoning
- [ ] PostgreSQL selected over MySQL/SQLite/MongoDB with reasoning
- [ ] Goose selected over other migration tools with reasoning
- [ ] Repository Pattern selected over Active Record/ORM/no abstraction with reasoning
- [ ] Manual DI selected over Service Locator/DI Framework with reasoning
- [ ] Strategy Pattern selected for shortcode generation with reasoning
- [ ] Each trade-off explicitly documented

### References

- See [ADR](#adr--architecture-decision-records) for detailed decision records
- See [Development Roadmap](#development-roadmap) for implementation order

---

## Project Structure

### Complete Folder Tree

```
url-shortener/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point: composition root
├── internal/
│   ├── config/
│   │   └── config.go                  # Config struct + Load() function
│   ├── logger/
│   │   └── logger.go                  # slog logger setup (development/prod)
│   ├── domain/
│   │   ├── url.go                     # URL entity, ShortCode type, status
│   │   └── errors.go                  # Domain sentinel errors (ErrURLNotFound, etc.)
│   ├── storage/
│   │   └── storage.go                 # PostgreSQL connection, Goose auto-migrate
│   ├── repository/
│   │   ├── repository.go              # URLRepository interface
│   │   └── postgres/
│   │       └── url.go                 # PostgreSQL implementation of URLRepository
│   ├── shortener/
│   │   └── shortener.go               # Shortener interface + RandomShortener
│   ├── validator/
│   │   └── validator.go               # ValidateURL, ValidateShortCode, ValidateID
│   ├── service/
│   │   ├── service.go                 # URLService interface
│   │   └── url_service.go            # URLService implementation
│   ├── handler/
│   │   ├── handler.go                 # BaseHandler (shared service + logger)
│   │   └── url.go                     # CreateURL, GetURL, Redirect, Update, Delete, Stats
│   ├── middleware/
│   │   └── middleware.go              # Logging, Request ID, Recovery middleware
│   ├── router/
│   │   └── router.go                  # HTTP route setup with http.ServeMux
│   ├── error/
│   │   └── error.go                   # AppError type (Code, Message, StatusCode)
│   └── health/
│       └── handler.go                 # Health check handler
├── migrations/
│   ├── 000001_create_urls_table.up.sql
│   ├── 000001_create_urls_table.down.sql
│   ├── 000002_add_indexes.up.sql
│   └── 000002_add_indexes.down.sql
├── docs/
│   └── software-design-document.md    # Complete architecture document
├── .env.example                       # Example environment variables
├── .gitignore                         # Go + Docker ignore patterns
├── .dockerignore                      # Docker build context exclusions
├── Dockerfile                         # Multi-stage build (builder + runtime)
├── docker-compose.yml                 # App + PostgreSQL for local dev
├── Makefile                           # build, test, fmt, vet, clean targets
├── go.mod                             # Go module definition
├── go.sum                             # Go module checksums
└── README.md                          # This document
```

### Package Responsibilities

**`cmd/server/`** — The entry point. Contains `main.go` which wires all dependencies (config, logger, storage, repository, service, handlers, router, middleware) and starts the HTTP server with `http.ListenAndServe`. This package has zero business logic — its sole responsibility is composition.

**`internal/config/`** — Defines the `Config` struct and a `Load() (*Config, error)` function. Reads all configuration from environment variables with sensible defaults. The `internal/` boundary prevents external packages from importing configuration.

**`internal/logger/`** — Initializes and returns a `*slog.Logger` configured for the environment. Uses `slog.NewTextHandler` for development and `slog.NewJSONHandler` for production. The logger is the first component that `main.go` creates.

**`internal/domain/`** — The purest package in the codebase. Contains the `URL` struct, `ShortCode` type, `Status` type, and sentinel errors (`ErrURLNotFound`, `ErrURLAlreadyDeleted`, `ErrShortCodeCollision`). Has zero imports from other `internal/` packages or external packages — only standard library.

**`internal/storage/`** — Manages the PostgreSQL connection using `database/sql` with the `pgx` driver. Creates the `*sql.DB` pool and auto-runs Goose migrations on initialization. Exposes `*sql.DB` via a `DB()` method for the repository to use.

**`internal/repository/`** — Defines the `URLRepository` interface with all data access method signatures. The concrete implementation lives in `repository/postgres/`. The service layer imports only this interface package.

**`internal/repository/postgres/`** — Implements `URLRepository` using `database/sql` with parameterized queries. Maps database rows to domain entities. Translates database errors (e.g., unique constraint violation) to application-level errors before returning to the service layer.

**`internal/service/`** — Contains the `URLService` interface and its implementation `urlService`. All business logic lives here: URL validation delegation, shortcode generation orchestration, redirect count management, and error wrapping. Has no knowledge of HTTP or databases.

**`internal/handler/`** — HTTP handlers for each endpoint. Handlers parse requests, call service methods, format responses, and set status codes. Handlers are intentionally thin — all business logic is delegated to the service.

**`internal/middleware/`** — Three middleware functions: logging (every request, method, path, status, duration, request_id), request ID (UUID v4 generation per request, stored in context, added to response headers), and recovery (catch panics, log them, return 500 with generic message).

**`internal/router/`** — Configures `http.ServeMux` with all route bindings. Maps HTTP method + path patterns to handler methods.

**`internal/shortener/`** — The `Shortener` interface + `RandomShortener` implementation. Uses `crypto/rand` for randomness and base62 encoding. The interface allows adding strategies like `HashShortener` or `SequentialShortener` in the future.

**`internal/validator/`** — Pure validation functions (`ValidateURL`, `ValidateShortCode`, `ValidateID`). No side effects, no I/O, no external dependencies. Validation uses `net/url` for URL parsing, not regex.

**`internal/error/`** — The `AppError` struct with `Code`, `Message`, and `StatusCode` fields. Constructor functions for common error types (`NewValidationError`, `NewNotFoundError`, `NewConflictError`, `NewGoneError`, `NewInternalError`). Ensures consistent error responses across all endpoints.

**`internal/health/`** — Health check handler that verifies database connectivity via a simple query (`SELECT 1`) and returns the service status including a timestamp.

**`migrations/`** — Goose SQL migration files. Each migration has an `.up.sql` (apply changes) and `.down.sql` (rollback changes). Numbered sequentially (`000001`, `000002`, etc.) for ordered, deterministic application.

**`docs/`** — Swagger/OpenAPI specification file (`swagger.yaml`) and the software design document referenced throughout this README.

### Dependency Direction Rules (Enforced)

```
cmd/server/          (most volatile, depends on everything)
    ↓ depends on
handler/             (depends on service only)
    ↓ depends on
service/             (depends on repository interface only)
    ↓ depends on
repository/          (depends on domain only)
    ↓ implements
domain/              (depends on nothing)
```

**Import cycle rule:** No package may import a package that imports it. The dependency graph must be a Directed Acyclic Graph (DAG). `go vet` catches import cycles automatically.

**Internal boundary:** The `internal/` directory enforces encapsulation at the Go build level. No external package can import anything inside `internal/`. This prevents consumers from depending on internal implementation details.

### Best Practices

- Every package directory has a clear, singular responsibility
- Internal packages never import cmd/ (breaking the encapsulation boundary)
- The domain package has zero external dependencies — it is the most stable and importable package
- All SQL is parameterized — no string interpolation of user input into queries
- All functions accept `context.Context` as their first parameter for timeout and cancellation support

### Common Mistakes

- **Circular imports**: If `service` imports `handler` and `handler` imports `service`, Go will refuse to compile with `import cycle not allowed`. Always draw the dependency graph before writing code.
- **Placing repository interface in the postgres sub-package**: The repository interface must be importable by the service layer. Place it in `repository/` (parent), not `repository/postgres/`.
- **Importing `database/sql` in the service layer**: The service should never know about `database/sql`. If you see `import "database/sql"` in any file under `internal/service/`, refactor immediately.
- **Importing `net/http` in anything except `handler/`**: Handlers own HTTP types. Nothing else should import `net/http`.

### Senior Tips

Run `go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./...` periodically to inspect the import graph. If you see `internal/handler` importing `internal/repository/postgres`, you have a dependency inversion violation. Fix it by introducing the repository interface in `internal/repository/` and having the service depend on that interface instead.

### Checklist

- [ ] `cmd/server/main.go` entry point defined
- [ ] `internal/config/config.go` config package created
- [ ] `internal/logger/logger.go` logger package created
- [ ] `internal/domain/url.go` URL entity defined
- [ ] `internal/domain/errors.go` domain errors defined
- [ ] `internal/storage/storage.go` database connection defined
- [ ] `internal/repository/repository.go` URLRepository interface defined
- [ ] `internal/repository/postgres/url.go` PostgreSQL implementation created
- [ ] `internal/shortener/shortener.go` strategy interface + RandomShortener created
- [ ] `internal/validator/validator.go` validation functions defined
- [ ] `internal/service/service.go` URLService interface defined
- [ ] `internal/service/url_service.go` URLService implementation created
- [ ] `internal/handler/handler.go` BaseHandler created
- [ ] `internal/handler/url.go` all HTTP handlers created
- [ ] `internal/middleware/middleware.go` logging, request ID, recovery middleware
- [ ] `internal/router/router.go` HTTP route configuration
- [ ] `internal/error/error.go` AppError type defined
- [ ] `internal/health/handler.go` health check handler created
- [ ] `migrations/` directory with up/down SQL files
- [ ] `docs/` directory for Swagger/OpenAPI spec
- [ ] `internal/` boundary enforced (no external imports)
- [ ] No circular imports in the project
- [ ] Dependency graph is acyclic and follows correct direction

### References

- See [Architecture](#architecture) for layer responsibilities and dependency direction
- See [Database Design](#database-design) for storage layer details
- See [Development Roadmap](#development-roadmap) for build order

---

## Database Design

### Schema

```sql
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL,
    redirect_count BIGINT DEFAULT 0 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

### Column Specifications

| Column | Type | Constraint | Purpose |
|---|---|---|---|
| `id` | `SERIAL` | `PRIMARY KEY` | Internal auto-increment identifier. Never exposed to clients. |
| `short_code` | `VARCHAR(255)` | `NOT NULL UNIQUE` (partial index) | Public-facing 8-char base62 token. 255 chars allows flexibility if shortcode length changes. |
| `original_url` | `TEXT` | `NOT NULL` | Full destination URL. `TEXT` has no length limit in PostgreSQL and performs identically to `VARCHAR(n)`. |
| `redirect_count` | `BIGINT` | `DEFAULT 0 NOT NULL` | Counter of successful redirects. `BIGINT` (int64) prevents overflow at 9.2 quintillion clicks. |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Immutable creation timestamp. Stored in UTC. |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Updated on every write (INSERT, UPDATE, DELETE). |
| `deleted_at` | `TIMESTAMPTZ` | Nullable | Soft delete marker. NULL = active, non-NULL = deleted. |

### Indexes

| Index | Definition | Purpose |
|---|---|---|
| `PK_urls` | `PRIMARY KEY (id)` | Fast internal ID lookups (rare, for admin/debugging) |
| `idx_urls_short_code` | `UNIQUE INDEX ON urls(short_code) WHERE deleted_at IS NULL` | The critical index — every redirect and API lookup uses this. Partial index allows shortcode reuse after soft delete. |
| `idx_urls_original_url_pattern` | `CREATE INDEX idx_urls_original_url ON urls(original_url text_pattern_ops)` | Optional — enables prefix matching for "does this URL already exist?" queries |
| `idx_urls_created_at` | `CREATE INDEX idx_urls_created_at ON urls(created_at DESC)` | Time-based queries for analytics and admin views |

### Partial Unique Index (Critical Design Decision)

```sql
CREATE UNIQUE INDEX idx_urls_short_code ON urls(short_code) WHERE deleted_at IS NULL;
```

This partial index enforces uniqueness only for active (non-deleted) URLs. This allows soft-deleted shortcodes to be reused — when a user creates a new URL with a previously deleted shortcode, the index check passes because the existing row has `deleted_at IS NOT NULL`, so it is excluded from uniqueness enforcement.

**Why not a regular UNIQUE constraint?** A regular `UNIQUE(short_code)` constraint would prevent reuse of soft-deleted shortcodes entirely. The partial index elegantly solves this by excluding deleted URLs from the uniqueness constraint.

### Constraints

```sql
ALTER TABLE urls ADD CONSTRAINT chk_original_url CHECK (original_url IS NOT NULL AND original_url != '');
ALTER TABLE urls ADD CONSTRAINT chk_redirect_count CHECK (redirect_count >= 0);
ALTER TABLE urls ADD CONSTRAINT chk_short_code_length CHECK (LENGTH(short_code) = 8);
```

- `chk_original_url`: Ensures no empty URLs are stored (defense in depth beyond application validation)
- `chk_redirect_count`: Prevents negative redirect counts
- `chk_short_code_length`: Ensures shortcodes are always 8 characters (can be relaxed later if shortcode length changes)

### Normalization

The schema is in **First Normal Form (1NF)** — atomic values, no repeating groups. It is intentionally not normalized further because:

1. **Single table avoidance of JOINs**: The most common operation (find URL by shortcode for redirect) is a single-table lookup. Additional tables would require JOINs for no benefit.
2. **Natural evolution**: If features like Users, Campaigns, or Tags are added, normalization occurs naturally when those entities are introduced as separate tables with foreign keys.
3. **Read-heavy workload**: URL shorteners are overwhelmingly read-heavy (millions of redirects per create). Denormalization for read performance is acceptable and even preferred at this scale.

### Migration Strategy

**Goose** manages database migrations with numbered SQL files.

**Naming convention:**
```
migrations/
  000001_create_urls_table.up.sql     # Apply: create table + constraints
  000001_create_urls_table.down.sql   # Rollback: drop table
  000002_add_indexes.up.sql           # Apply: add indexes + partial unique index
  000002_add_indexes.down.sql         # Rollback: drop indexes
  000003_add_shortcode_length_check.up.sql  # Future
```

**How migrations work:**
1. Goose tracks applied migrations in an internal `goose_db_version` table
2. On startup, `Storage` auto-runs `goose.Up(db, "migrations")` to apply pending migrations
3. Each migration is a SQL file — readable, reviewable, and version-controlled
4. Down migrations enable rollback for development and emergency recovery

**Production migration strategy:**
- Migrations are applied on application startup (auto-apply)
- In CI/CD, run migrations as a separate step before deploying the application
- Always write `.down.sql` before `.up.sql` for every migration
- Test migration rollbacks manually before relying on them

### Future Scalability

For very high scale (100M+ URLs), consider:
1. **Partitioning**: Partition the `urls` table by `created_at` range for better query performance and easier archival
2. **Read replicas**: Route redirect (read) traffic to PostgreSQL read replicas for horizontal read scaling
3. **Connection pooling**: Use `pgxpool` instead of raw `sql.DB` for more sophisticated connection management at scale
4. **Caching layer**: Add Redis for frequently accessed shortcodes (see [Future Improvements](#future-improvements))

### Mermaid ER Diagram

```mermaid
erDiagram
    urls {
        serial id PK "auto-increment, internal only"
        varchar short_code UK "unique WHERE deleted_at IS NULL"
        text original_url "unlimited length, NOT NULL"
        bigint redirect_count ">= 0, DEFAULT 0"
        timestamptz created_at "UTC, immutable"
        timestamptz updated_at "UTC, updated on every write"
        timestamptz deleted_at "nullable, deletion marker"
    }
```

### Engineering Notes

The partial unique index is the most important database-level decision in this project. Without it, you cannot reuse deleted shortcodes — either the UNIQUE constraint blocks new URLs or you must manually delete the old row (losing the ability to soft-delete). The partial index elegantly handles both requirements simultaneously.

`TIMESTAMPTZ` (TIMESTAMP WITH TIME ZONE) stores timestamps in UTC internally. PostgreSQL converts from the session timezone on input and to UTC for storage. The application should always use UTC for all timestamps and convert to local time only at the presentation layer.

`SERIAL` (32-bit) is sufficient for the primary key. At 2.1 billion rows, you can revisit using `BIGSERIAL` for extreme scale. The shortcode (not the ID) is what clients interact with, so the internal ID is never exposed.

### Common Mistakes

- **Using `UNIQUE(short_code)` without the `WHERE deleted_at IS NULL` clause**: This prevents shortcode reuse after soft deletion, defeating the business rule BR-05.
- **Using `VARCHAR(8)` for short_code**: If you later want to support longer or shorter shortcodes, you will need to alter the column type. `VARCHAR(255)` is intentionally loose to allow future flexibility.
- **Using `NOW()` for `updated_at` instead of setting it explicitly on UPDATE**: If you rely on `DEFAULT NOW()` for updates, the timestamp won't change unless you explicitly set it in your UPDATE statement. Either handle `updated_at` explicitly or use a PostgreSQL trigger.
- **Not using `text_pattern_ops` for the optional original_url index**: Without this operator class, a B-tree index on `TEXT` cannot support `LIKE` pattern matching efficiently.
- **Storing `NULL` in `deleted_at` instead of using it as a deletion marker**: Some databases treat NULL timestamps differently. In PostgreSQL, NULL means the URL is active — this is explicit and intentional.

### Senior Tips

Test your partial unique index manually: first create a URL with shortcode "abc123de", soft-delete it, then create a new URL with the same shortcode. Without the partial index, this second CREATE would fail with a unique constraint violation. With it, the second CREATE succeeds.

### Checklist

- [ ] Schema defined with all 7 columns and correct types
- [ ] Primary key (id SERIAL) defined
- [ ] Unique constraint on short_code (partial index with deleted_at IS NULL)
- [ ] Index on short_code defined and optimized
- [ ] Check constraints defined (original_url not empty, redirect_count >= 0, short_code length = 8)
- [ ] Normalization level documented (1NF)
- [ ] Migration strategy (Goose) documented with file naming convention
- [ ] Mermaid ER diagram created
- [ ] Column explanations for each of the 7 columns
- [ ] Why SERIAL over BIGSERIAL explained
- [ ] Why TIMESTAMPTZ over TIMESTAMP explained
- [ ] Why partial unique index explained (shortcode reuse)
- [ ] Why TEXT over VARCHAR(n) explained
- [ ] Future scalability considerations documented

### References

- See [Domain Analysis](#domain-analysis) for business rules that DB enforces
- See [Validation Rules](#validation-rules) for application-layer validation
- See [ADR](#adr--architecture-decision-records) for PostgreSQL decision rationale
- See [Development Roadmap](#development-roadmap) for Phase 3 (Database)

---

## API Design

### REST Principles

The API follows RESTful conventions:

- **Resources are nouns**: `/urls`, `/urls/{shortcode}` — not verbs
- **HTTP methods map to operations**: GET (read), POST (create), PUT (update), DELETE (remove)
- **Versioned**: All endpoints prefixed with `/api/v1/` to allow future breaking changes
- **Stateless**: Each request contains all information needed to process it — no session state
- **Resource-oriented**: Responses contain the full resource representation, not just IDs

### Base URL

```
http://localhost:8080/api/v1/
```

The redirect endpoint uses the root path for brevity:
```
http://localhost:8080/r/{shortcode}
```

### API Reference

---

#### Create Short URL

**POST** `/api/v1/urls`

**Request Body:**
```json
{ "url": "https://example.com/very/long/path?query=param" }
```

| Field | Type | Required | Description |
|---|---|---|---|
| `url` | `string` | Yes | The original long URL to shorten |

**Success Response — `201 Created`:**
```json
{
  "id": 1,
  "short_code": "abc123de",
  "short_url": "https://short.io/abc123de",
  "original_url": "https://example.com/very/long/path?query=param",
  "redirect_count": 0,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

| Status | When | Body |
|---|---|---|
| `201 Created` | URL created successfully | URL object as above |
| `400 Bad Request` | Missing `url` field or invalid URL | Error object |
| `422 Unprocessable Entity` | URL exceeds 2048 characters | Error object |

**Error Response — `400 Bad Request`:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "url is required",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

#### Redirect

**GET** `/r/{shortcode}`

**Path Parameters:**

| Name | Type | Required | Description |
|---|---|---|---|
| `shortcode` | `string` | Yes | 8-character shortcode |

**Success Response — `302 Found`:**
```
HTTP/1.1 302 Found
Location: https://example.com/very/long/path
Cache-Control: no-store
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```
- **No body** — the browser/follows the redirect automatically
- **`Cache-Control: no-store`** prevents caching of redirect responses
- **`302 Found`** is used (not 301) because 301s are cached aggressively by browsers, preventing redirect count tracking and future destination changes

**Error Responses:**

| Status | When | Body |
|---|---|---|
| `404 Not Found` | Shortcode does not exist | Error object |
| `410 Gone` | Shortcode was soft-deleted | Error object |
| `400 Bad Request` | Shortcode format invalid | Error object |

---

#### Get URL Info

**GET** `/api/v1/urls/{shortcode}`

**Path Parameters:**

| Name | Type | Required | Description |
|---|---|---|---|
| `shortcode` | `string` | Yes | 8-character shortcode |

**Success Response — `200 OK`:**
```json
{
  "id": 1,
  "short_code": "abc123de",
  "original_url": "https://example.com/very/long/path",
  "redirect_count": 42,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T12:00:00Z"
}
```

**Error Responses:** `404 Not Found`, `400 Bad Request`

---

#### Update URL

**PUT** `/api/v1/urls/{shortcode}`

**Request Body:**
```json
{ "url": "https://new-destination.com/new-path" }
```

**Success Response — `200 OK`:**
```json
{
  "id": 1,
  "short_code": "abc123de",
  "original_url": "https://new-destination.com/new-path",
  "redirect_count": 42,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T13:00:00Z"
}
```

**Error Responses:**

| Status | When |
|---|---|
| `400 Bad Request` | Missing or invalid URL field |
| `404 Not Found` | Shortcode does not exist (including soft-deleted) |

---

#### Delete URL (Soft Delete)

**DELETE** `/api/v1/urls/{shortcode}`

**Success Response — `204 No Content`:** (empty body)

**Error Responses:** `404 Not Found` (including already-deleted URLs)

**Note:** Successful deletion does not return a JSON body. This follows REST convention for DELETE — no content on success.

---

#### Get Statistics

**GET** `/api/v1/urls/{shortcode}/stats`

**Success Response — `200 OK`:**
```json
{
  "short_code": "abc123de",
  "redirect_count": 42,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T13:00:00Z"
}
```

**Error Responses:** `404 Not Found`

---

#### Health Check

**GET** `/health`

**Success Response — `200 OK`:**
```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

No path parameters, no body on success (for `GET`). Used by load balancers and orchestrators for liveness/readiness probes.

---

#### Error Response Format (Consistent)

All error responses follow this structure:
```json
{
  "error": {
    "code": "<MACHINE_READABLE_CODE>",
    "message": "<Human-readable description>",
    "request_id": "<UUID v4>"
  }
}
```

| Error Code | HTTP Status | When |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Input validation failed |
| `NOT_FOUND` | 404 | Resource not found |
| `ALREADY_DELETED` | 404 | Resource was soft-deleted |
| `CONFLICT` | 409 | Resource conflict |
| `GONE` | 410 | Redirect target was deleted |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

The `request_id` is the unique UUID assigned by the middleware to the incoming request. It enables log correlation between the application's log records and the error response returned to the caller.

### Swagger Notes

- Use hand-written OpenAPI 3.0 YAML in `docs/swagger.yaml` — no external annotation tools required
- Tag endpoints as `/urls` (CRUD operations) and `/redirect` (302 redirect endpoint)
- Include example request/response bodies for every endpoint at every status code
- The Swagger UI should be accessible at `/swagger/index.html` in development
- Define reusable components for shared schemas (`URL`, `ErrorResponse`, `StatsResponse`)
- Document the 302 redirect specially — it has no JSON body but includes the `Location` header

### Checklist

- [ ] POST /api/v1/urls documented with full request/response/error examples
- [ ] GET /r/{shortcode} (redirect) documented with 302 response
- [ ] GET /api/v1/urls/{shortcode} documented
- [ ] PUT /api/v1/urls/{shortcode} documented
- [ ] DELETE /api/v1/urls/{shortcode} documented
- [ ] GET /api/v1/urls/{shortcode}/stats documented
- [ ] GET /health documented
- [ ] All status codes listed for each endpoint
- [ ] Example JSON bodies for success and error cases
- [ ] Reason for 302 instead of 301 documented
- [ ] Error response format standardized and documented
- [ ] REST principles documented
- [ ] Base URL documented

### References

- See [Functional Requirements](#software-requirements) for FR traceability
- See [Business Rules](#domain-analysis) for rule enforcement at API level
- See [Development Roadmap](#development-roadmap) for endpoint implementation order
- See [Swagger](#testing-strategy) for documentation implementation

---

## Validation Rules

### URL Validation (Application Layer)

Performed by `internal/validator/validator.go` using `net/url` package parsing (not regex).

| Rule | Error | Rationale |
|---|---|---|
| `url` field must be present and non-empty | `ErrEmptyURL` | Prevents storing empty or whitespace-only URLs |
| URL must parse successfully with `url.Parse` | `ErrInvalidURL` | Catches malformed URLs that cannot be processed |
| URL scheme must be `http` or `https` | `ErrInvalidURLScheme` | Prevents non-HTTP URLs (ftp://, javascript:, mailto:) |
| URL must have a non-empty host | `ErrInvalidURLHost` | A URL without a host is useless for redirection |
| URL length must not exceed 2048 characters | `ErrURLTooLong` | 2048 is the browser URL bar limit |

**Implementation note:** `strings.TrimSpace(url) == ""` check catches whitespace-only strings before parsing. `url.Parse` is called on the original (non-trimmed) URL to preserve the exact user input.

### ShortCode Validation

Performed by `internal/validator/validator.go`.

| Rule | Error | Rationale |
|---|---|---|
| ShortCode must be exactly 8 characters | `ErrInvalidShortCodeLength` | Fixed length simplifies generation and validation |
| ShortCode must contain only base62 characters (a-z, A-Z, 0-9) | `ErrInvalidShortCodeChars` | Prevents URL injection or encoding issues |

### Business Validation (Service Layer)

| Rule | Error | Rationale |
|---|---|---|
| URL must exist for the given shortcode | `ErrURLNotFound` | Cannot operate on a nonexistent URL |
| URL must be active (deleted_at is nil) | `ErrURLAlreadyDeleted` (410) | Soft-deleted URLs cannot be redirected or updated |
| Shortcode collision after max attempts | `ErrShortCodeCollision` | Auto-generation failed (extremely rare) |

### Database Validation (Defense in Depth)

| Rule | Mechanism | Rationale |
|---|---|---|
| ShortCode uniqueness | `UNIQUE INDEX ... WHERE deleted_at IS NULL` (partial) | Prevents duplicate shortcodes even under race conditions |
| Original URL not empty | `CHECK (original_url IS NOT NULL AND original_url != '')` | Catches invalid URLs that bypass application validation |
| Redirect count non-negative | `CHECK (redirect_count >= 0)` | Prevents counter corruption |
| ShortCode length fixed | `CHECK (LENGTH(short_code) = 8)` | Enforces fixed-length contract at DB level |

### Engineering Notes

Validation operates at two levels as defense in depth:

1. **Application layer**: Fast, descriptive error messages for the client. Catches obvious errors before any database operation.
2. **Database layer**: The last line of defense. Catches race conditions (e.g., two concurrent creates generating the same shortcode) and corrupt data that bypasses application logic.

The service layer does not duplicate validation — it delegates to the validator package. This follows the DRY principle and ensures validation logic is defined in one place.

### Common Mistakes

- **Using regex for URL validation**: `net/url.Parse` is the correct approach. Regex URL validators are notoriously error-prone and cannot match the correctness of Go's standard URL parser.
- **Validating once and assuming it stays valid**: In concurrent systems, the database constraint catches validation that was valid at read time but violated at write time (e.g., duplicate shortcode generation).
- **Not trimming whitespace**: A URL like `"  https://example.com  "` should be rejected, not silently trimmed and stored. The validator rejects whitespace-only strings and strings that become invalid after trimming.

### Senior Tips

The validator package is the simplest package in this project. It has no dependencies, no state, and no side effects. If your validator grows beyond 100 lines, consider whether it belongs in the service layer instead (if it needs to call other services) or stays pure (if it only needs `net/url`).

### Checklist

- [ ] URL scheme validation (http/https only)
- [ ] URL host validation (non-empty)
- [ ] URL length validation (max 2048 chars)
- [ ] Whitespace-only URL rejection
- [ ] ShortCode length validation (exactly 8 chars)
- [ ] ShortCode character validation (base62 only)
- [ ] ShortCode uniqueness checked (via repository)
- [ ] URL existence checked before update/delete/redirect
- [ ] Soft-deleted URLs return appropriate error (410)
- [ ] Database constraints mirror application validation (defense in depth)

### References

- See [Business Rules](#software-requirements) for BR-01 to BR-10
- See [Domain Analysis](#domain-analysis) for invariants
- See [Service Design](#development-roadmap) for where validation is applied

---

## Error Handling

### Error Types

The project defines two tiers of errors:

**Tier 1 — Application Errors (`internal/error/error.go`)**

| Error Type | HTTP Status | When Returned |
|---|---|---|
| `AppError` | Configurable | Base type for application-level errors |
| `ValidationError` | 400 | Invalid user input |
| `NotFoundError` | 404 | Resource doesn't exist |
| `ConflictError` | 409 | Duplicate resource creation |
| `GoneError` | 410 | Resource was soft-deleted |
| `InternalError` | 500 | Unexpected server failure |

**Tier 2 — Domain Sentinel Errors (`internal/domain/errors.go`)**

| Sentinel Error | When Used |
|---|---|
| `ErrURLNotFound` | URL with given shortcode does not exist |
| `ErrURLAlreadyDeleted` | URL exists but has been soft-deleted |
| `ErrShortCodeCollision` | Auto-generation exhausted max attempts |
| `ErrInvalidShortCode` | Shortcode format is invalid |
| `ErrInvalidURL` | Original URL format is invalid |

### Error Wrapping Chain

Errors propagate through layers with context wrapping using `%w`:

```
database/sql error
  → repository: fmt.Errorf("repository: failed to get URL: %w", err)
    → service: fmt.Errorf("service: failed to get URL by shortcode: %w", err)
      → handler: detects sentinel error with errors.Is, formatsAppError for HTTP response
```

**The `errors.Is` chain ensures that:**
- A handler can detect `ErrURLNotFound` even though it was wrapped by the service layer and the repository layer
- Context is preserved at every layer without losing the original error identity
- Tests can assert on sentinel errors using `errors.Is(err, ErrURLNotFound)` directly

### HTTP Error Response Format

```json
{
  "error": {
    "code": "<MACHINE_READABLE_CODE>",
    "message": "<Human-readable description>",
    "request_id": "<UUID v4>"
  }
}
```

**How it works:**
1. The handler catches an error from the service
2. The handler checks `errors.Is(err, ErrURLNotFound)` to detect sentinel errors
3. If the error is an `*AppError`, use its `Code`, `Message`, and `StatusCode` directly
4. If the error is an unknown/unexpected error, return 500 with `code: "INTERNAL_ERROR"` and a generic `message`
5. The `request_id` is read from `r.Context().Value(requestIDKey)` — set by request ID middleware
6. The `request_id` is always included in error responses for log correlation

### Logging Strategy for Errors

Every error returned to a client is also logged with:
- `level`: `error` for 4xx/5xx responses, `info` for successful requests
- `request_id`: for log correlation
- `method`: HTTP method
- `path`: request path
- `status`: HTTP status code
- `error_code`: machine-readable error code
- `error_message`: the error message (never the full URL or PII)
- `stack_trace`: only in development mode (`APP_ENV=development`)

### Engineering Notes

The error wrapping pattern `fmt.Errorf("context: %w", err)` is critical. Without `%w`, the `errors.Is` chain breaks and handlers cannot detect sentinel errors. Each layer adds context that helps debugging but preserves the original error for type-based handling.

In production mode, the error `message` returned to the client should be generic ("Something went wrong"). In development mode, the message can be more specific. The `stack_trace` field is only included in development mode responses.

### Common Mistakes

- **Forgetting `%w` when wrapping errors**: `fmt.Errorf("repo: %v", err)` loses the error chain. `errors.Is` will not work across layers. Always use `%w`.
- **Returning raw database errors in the HTTP response**: This leaks internal information (table names, SQL states) to clients and is a security risk. Map all database errors to `AppError` before they leave the repository layer.
- **Using `errors.New` instead of sentinel `errors.Is` for comparison**: Define sentinel errors as package-level `var` declarations so they support `errors.Is` across wrapping boundaries.
- **Never returning 500 for user errors**: Always return the correct status code. A validation error should never return 500. A not-found should never return 500. Reserve 500 for truly unexpected internal failures.

### Senior Tips

Create a helper in the handler package called `writeError(w http.ResponseWriter, r *http.Request, err error)` that maps errors to HTTP responses uniformly. This eliminates duplicate error-handling code across all handler methods and ensures consistency. The helper should:
1. Check `errors.Is` for each sentinel error
2. Log the error with appropriate fields
3. Write the JSON error response with the correct status code
4. Include request_id

### Checklist

- [ ] AppError struct defined (Code, Message, StatusCode)
- [ ] Error constructor functions (NewValidationError, NewNotFoundError, NewConflictError, NewGoneError, NewInternalError)
- [ ] Domain sentinel errors defined (ErrURLNotFound, ErrURLAlreadyDeleted, ErrShortCodeCollision)
- [ ] Error wrapping with %w at every layer boundary
- [ ] errors.Is used in handlers to detect sentinel errors
- [ ] Consistent error response format (code, message, request_id)
- [ ] 500 responses mask internal error details in production
- [ ] Stack traces only in development mode
- [ ] No raw database errors exposed to clients
- [ ] No user-facing errors use %v instead of %w

### References

- See [Logging Strategy](#logging-strategy) for error logging format
- See [Validation Rules](#validation-rules) for error codes in validation
- See [Handler Design](#development-roadmap) for where error handling is implemented
PART6EOF
echo "Part 7 written ($(wc -l < /run/media/mmk/6A2447ED2447BAB9/projects/url-shortener/README.md) lines)"

---

## Logging Strategy

### Why log/slog

`log/slog` is the standard library structured logger introduced in Go 1.21. It provides structured JSON output for production, human-readable output for development, log levels (Debug/Info/Warn/Error), and pluggable handlers — all with zero external dependencies.

**Why not zap or zerolog:** They add external dependencies. For an educational project emphasizing standard library usage, slog is the right choice. zap/zerolog can be adopted later if profiling shows slog is a bottleneck.

**Why not the deprecated `log` package:** The standard `log` package only supports unstructured text logging. slog provides structured logging out of the box, essential for production observability and log aggregation (ELK, Datadog, Loki).

### Log Levels

| Level | When to Use | Example |
|---|---|---|
| `Debug` | Verbose diagnostic info, development only | SQL query text, request body parsing |
| `Info` | Normal operational events | Request completed, URL created, redirect served |
| `Warn` | Unexpected but recoverable situations | Request to non-standard port URL |
| `Error` | Failed operations that can be recovered from | Repository query failed with database unavailable |

### Structured Logging

Production logs are JSON with consistent field naming:

```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "info",
  "msg": "request completed",
  "method": "POST",
  "path": "/api/v1/urls",
  "status": 201,
  "duration_ms": 12,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Request ID

Every request gets a unique UUID v4 `request_id` generated by the middleware. It is included in every log line for the request's lifecycle and in every HTTP error response for correlation.

### What Should Be Logged

- Every incoming request: method, path, status code, duration, request_id
- Every URL creation: shortcode (never full URL — may contain sensitive query params)
- Every redirect: shortcode, truncated original_url hash, redirect count after increment
- Every error: error_code, request_id, contextual field (shortcode), stack_trace in development only
- Server start/shutdown events, configuration at startup (sanitized — no secrets)

### What Should NEVER Be Logged

- **Full original URLs** — they may contain sensitive query parameters (API keys, tokens, PII)
- **Request bodies with sensitive data** — log metadata, not raw content
- **Database credentials / DSN strings** — never log connection details
- **Internal IP addresses or PII** — anonymize or omit user IP addresses

### Configuration

Development mode: text handler, level Debug. Production mode: JSON handler, level Info. Switch based on `APP_ENV`.

### Engineering Notes

Use `slog.Info("URL created", "shortcode", code, "original_url_hash", hash)` with named key-value pairs — never string concatenation. Use `slog.With` to create child loggers pre-bound with request_id to avoid passing it to every log call.

### Common Mistakes

- Logging at Debug level in production — Debug logs are extremely verbose and impact performance
- Logging the full DSN string — a classic security mistake containing database credentials
- Forgetting to set request_id in the logging context — makes log correlation impossible in distributed tracing

### Senior Tips

Use `slog.New(slog.NewJSONHandler(io.Discard, ...)` to temporarily disable logging in tests that don't need log output. In production, point slog's JSON stdout output to a file or pipe to a log shipper (Fluentd, Datadog agent, Loki).

### Checklist

- [ ] log/slog selected and justified over zap/zerolog/logrus
- [ ] Log levels defined (Debug, Info, Warn, Error)
- [ ] Structured JSON logging for production
- [ ] Human-readable logging for development
- [ ] Request ID generated per request and propagated through context
- [ ] Request ID included in every log line
- [ ] Request ID included in every error response
- [ ] What should be logged is defined (requests, URL creation, redirects, errors)
- [ ] What should NEVER be logged is defined (full URLs, secrets, PII, DSN)
- [ ] Logger configuration (development vs production) documented
- [ ] slog handler setup documented
- [ ] Child logger pattern documented (slog.With)

---

## Design Patterns

### Repository Pattern

**What:** Abstraction layer between business logic and data storage. The service depends on `URLRepository` interface, not a concrete database implementation.

**How used:**
```go
// repository/repository.go
type URLRepository interface {
    Create(ctx context.Context, url *domain.URL) (*domain.URL, error)
    GetByShortCode(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
    GetAll(ctx context.Context) ([]domain.URL, error)
    GetByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error)
    Update(ctx context.Context, code domain.ShortCode, url *domain.URL) (*domain.URL, error)
    SoftDelete(ctx context.Context, code domain.ShortCode) error
    IncrementRedirectCount(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
}
```

The concrete implementation in `internal/repository/postgres/url.go` uses `database/sql` with parameterized queries. The service imports only the interface.

**Why selected:** Enables unit testing with mock repositories, allows swapping storage backends, decouples business logic from SQL and connection details.

**Why not Active Record:** Coupling domain logic to persistence violates Single Responsibility Principle.

**Why not Data Mapper with GORM:** ORM adds magic behavior hard to debug, implicit, and less transparent than hand-written SQL.

### Dependency Injection

**What:** Dependencies provided from the outside, not created inside the consumer.

**How used:**
```go
// cmd/server/main.go
repo := postgres.NewURLRepository(db)
svc := service.NewURLService(repo)
handler := handler.NewURLHandler(svc)
```

**Why selected:** Makes testing trivial (pass mock repository), makes configuration explicit, follows Go's preference for explicit over implicit.

**Why not Service Locator:** Hidden dependencies make the code harder to trace and test.

**Why not DI frameworks (Uber/dig):** Unnecessary complexity for a project of this scale. Manual DI is clear and debuggable.

### Strategy Pattern

**What:** Family of interchangeable algorithms behind a common interface.

**How used:**
```go
type Shortener interface {
    Generate() (domain.ShortCode, error)
}

type RandomShortener struct { length int; alphabet string }
func (s *RandomShortener) Generate() (domain.ShortCode, error) {
    // crypto/rand + base62 encoding
}
```

**Why selected:** Shortcode generation algorithm is a natural candidate for swapping (random → hash → sequential). Keeps service independent of generation details. Easy to test with a deterministic shortener.

### Why These Three Patterns Together

Repository abstracts storage. DI wires dependencies. Strategy abstracts algorithms. Each solves a different problem and composes naturally. Do not force additional patterns — if a pattern adds complexity without clear benefit, skip it.

### Common Mistakes

- **Over-segmenting the repository interface**: If it has >10 methods, the implementing struct must implement all of them. Keep it focused.
- **Injected too many dependencies in service**: If the service constructor takes 8 parameters, it's hard to read. Consider grouping related dependencies into a config struct.
- **Making the repository too granular**: An interface with 50 methods couples the service to a bloated contract. A small interface (5-8 methods) is better.

### Senior Tips

Start with direct instantiation in main.go. Introduce DI containers only when wiring becomes unmanageable (typically at 10+ dependencies). The single most impactful pattern in this project is dependency inversion via interfaces — it enables testing with zero cost.

### Checklist

- [ ] Repository Pattern explained with interface and implementation
- [ ] Dependency Injection explained (manual, constructor-based)
- [ ] Strategy Pattern for shortcode generator explained
- [ ] Each pattern's rationale documented (why selected, why rejected)
- [ ] Project-specific examples provided

---

## SOLID Principles

### S — Single Responsibility Principle

Each type/module has exactly one reason to change.

**Examples in this project:**
- **Handler**: Parses HTTP, calls service, formats response. Does NOT validate business rules, does NOT write SQL.
- **Service**: Orchestrates business logic. Does NOT know about HTTP (`net/http`), does NOT write SQL (`database/sql`).
- **Repository**: Executes SQL, scans rows. Does NOT contain business rules.
- **Shortener**: Generates shortcodes only.
- **Validator**: Validates URLs only. No I/O, no state.

### O — Open/Closed Principle

Open for extension, closed for modification.

**Examples:**
- `URLRepository` interface is closed for modification but open for extension (`MockURLRepository`, `SQLiteURLRepository`, `InMemoryURLRepository` can be added without modifying service).
- `Shortener` interface allows adding `HashShortener`, `SequentialShortener` without modifying service.
- Middleware chain can be extended with new middleware without modifying existing handlers.

### L — Liskov Substitution Principle

Subtypes are substitutable for their base types without altering correctness.

**Examples:**
- Any `URLRepository` implementation (Postgres, Mock, SQLite) can be passed to the service.
- Any `Shortener` implementation (Random, Hash, Sequential) can be passed to the service.
- The service does not type-switch on implementation — it trusts the interface contract.

### I — Interface Segregation Principle

No client should depend on methods it does not use.

**Example:**
```go
// Good — focused interface for redirect handler only
type Redirector interface {
    GetByShortCode(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
    IncrementRedirectCount(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
}
```

### D — Dependency Inversion Principle

High-level modules should not depend on low-level modules. Both should depend on abstractions.

**Dependency graph in this project:**
```
Handler → Service (interface) → Repository (interface) → Storage
                                    ↑
                        PostgresRepository (detail)
```

- Handler depends on Service interface (abstraction), not concrete `urlService`
- Service depends on `URLRepository` interface (abstraction), not concrete `PostgresURLRepository`
- Repository interface does not depend on anything else
- `PostgresURLRepository` (detail) depends on `URLRepository` interface (abstraction)

**How to achieve:** Define interfaces in the package that consumes them. Inject concrete implementations at composition time (in `main.go`). Never use `sql.DB`, `pgx.Conn`, or `http.Request` in service or repository interface packages.

### Common Mistakes

- **Over-segmenting interfaces into micro-interfaces**: `URLReader`, `URLWriter`, `URLDeleter` — when a single `URLRepository` with 7 methods is cleaner and more practical.
- **Inverting the wrong dependency**: Dependence inversion means high-level (service) should not know about low-level (Postgres). PostgresRepository is still concrete — it just doesn't import the service.

### Checklist

- [ ] Single Responsibility: each layer/type has one clear reason to change
- [ ] Open/Closed: repository and shortener interfaces are extensible
- [ ] Liskov Substitution: any implementation can be swapped
- [ ] Interface Segregation: repository interface is focused, not fat
- [ ] Dependency Inversion: handler → service interface → repository interface
- [ ] Concrete project examples provided for each principle
- [ ] Violation anti-patterns documented

---

## Clean Code Principles

### Naming Conventions

| Rule | Good | Bad |
|---|---|---|
| Descriptive names | `GetByShortCode`, `ValidateURL` | `Get`, `GG` |
| Booleans start with is/has/can | `isValid`, `hasScheme`, `canRedirect` | `check`, `func1` |
| No unclear abbreviations | `originalURL`, `redirectCount` | `origURL`, `rCount` |
| Package names lowercase singular | `model`, `handler`, `service` | `Models`, `Handlers` |
| Interface names are nouns/roles | `URLRepository`, `Shortener` | `IURLRepository` (no `I` prefix in Go) |
| Meaningful test names | `TestCreateURL_ValidURL_Returns201` | `Test1`, `TestCreate` |

### Small Functions

- Every function should do **one thing** and do it well
- Functions should be under **30 lines** — extract helpers if longer
- Functions should have no more than **3-4 parameters** — use a config struct if more
- Functions should have clear input → output — no hidden side effects (except repository/database functions)

### Error Handling

- Every function that can fail returns `error` as the last return value
- Errors are values — check them explicitly, never ignore them
- Wrap errors with context: `fmt.Errorf("creating URL: %w", err)`
- Never pass `*http.Request` or `*http.ResponseWriter` into service or repository functions

### Package Responsibilities

Each package has exactly one reason to change. The dependency graph must be a DAG (no cycles):

```
cmd/ → handler/ → service/ → repository/ → domain/
```

### Configuration

All configuration values come from environment variables with sensible defaults. No hardcoded configuration values in business logic code. Configuration is loaded once at startup and passed to every component that needs it.

### Magic Numbers

| Number | Named Constant | Location |
|---|---|---|
| `8` (shortcode length) | `ShortCodeLength` | `internal/shortener/shortener.go` |
| `62` (base62 alphabet size) | `base62Alphabet` | `internal/shortener/shortener.go` |
| `2048` (max URL length) | `MaxURLLength` | `internal/validation/validator.go` |

### Comments

- Comments explain **why**, not **what**
- Every exported function has a Go doc comment (`// FunctionName does...`)
- No commented-out code — if it is not needed, delete it
- `TODO` or `FIXME` comments are tracked issues, not design documents

### Common Mistakes

- **Using `%v` instead of `%w`**: `fmt.Errorf("repo: %v", err)` loses the error chain. `errors.Is` will not work across layers.
- **Naming variables after their type**: `urlString`, `countInt` — Go infers types. Name by meaning: `originalURL`, `totalRedirects`.

### Senior Tips

A function that is easy to test is usually well-designed. If testing a function requires mocking 5 things, consider splitting it. Read "Clean Code" by Robert C. Martin, then apply principles pragmatically to Go idioms.

### Checklist

- [ ] Naming conventions documented and applied
- [ ] Small functions rule applied (under 30 lines, max 3-4 parameters)
- [ ] Error handling strategy documented (%w wrapping)
- [ ] Package responsibilities defined (one reason to change per package)
- [ ] Dependency direction documented
- [ ] Magic numbers replaced with named constants
- [ ] Go doc comments on all exported functions
- [ ] Configuration externalized (env vars, no hardcoded secrets)
- [ ] Comments explain "why" not "what"

---

## Security

### Input Validation

All user input is validated before processing:

- **URL validation**: `net/url.Parse` validates scheme and host. Rejects non-HTTP schemes (`ftp://`, `javascript:`, `mailto:`) that could be vectors for attacks.
- **Shortcode validation**: Confirmed to be exactly 8 base62 characters. Rejects any input that could be used for path traversal or injection.
- **Length limits**: Max URL length of 2048 characters prevents denial-of-service via absurdly long inputs.

### SQL Injection Prevention

All database queries use parameterized statements. The repository never uses string interpolation or `fmt.Sprintf` to build SQL queries.

```go
// GOOD — parameterized query
row := db.QueryRowContext(ctx, "SELECT * FROM urls WHERE short_code = $1", code)

// BAD — SQL injection vulnerability
row := db.QueryRow("SELECT * FROM urls WHERE short_code = '" + code + "'")
```

### Open Redirect Prevention

The service does not redirect to arbitrary URLs — only to URLs that were explicitly stored by an authenticated user (in future versions). For now, the original URL is stored as-is, but validation ensures it has a valid scheme and host, preventing `javascript:` or `data:` URI schemes that could execute code in a browser context.

### Sensitive Data Protection

- **Original URLs are not logged**: Full URLs may contain API keys, session tokens, or PII in query parameters. Log only a hashed or truncated version.
- **Shortcodes are opaque tokens**: They contain no encoded information and cannot be exploited to extract data.
- **No credentials in URLs**: The service does not proxy requests — it only stores and returns the original URL string. Credentials are never passed through the service.

### HTTP Security Headers (Future)

When the service is production-ready, apply these headers:

| Header | Value | Purpose |
|---|---|---|
| `X-Content-Type-Options` | `nosniff` | Prevent MIME type sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-XSS-Protection` | `1; mode=block` | Enable XSS filtering |
| `Strict-Transport-Security` | `max-age=63072000` | Enforce HTTPS |
| `Cache-Control` | `no-store` | Prevent caching of sensitive responses |

### Future Authentication

Authentication is out of scope for the initial release but planned. Future authentication will use JWT tokens (access + refresh) with the `Authorization: Bearer <token>` header. The service will support per-user URL ownership and access control. The repository interface will add a `GetAllByUser(ctx context.Context, userID string)` method without modifying existing methods.

### Checklist

- [ ] URL validation using net/url (not regex) — schemes restricted to http/https
- [ ] Shortcode format validation (8 chars, base62 only)
- [ ] URL length limit enforced (max 2048 chars)
- [ ] All SQL queries are parameterized (no string interpolation)
- [ ] No open redirect vulnerability (only stored URLs are redirected to)
- [ ] Full URLs NOT logged (only hashed/truncated versions)
- [ ] No credentials or secrets exposed in responses or logs
- [ ] Security headers planned for production deployment
- [ ] Authentication planned as future improvement (see Future Improvements)

### Common Mistakes

- **Using regex for URL validation**: Regex-based URL validators are notoriously error-prone. `net/url.Parse` is the correct approach in Go.
- **Allowing `javascript:` scheme URLs**: A URL like `javascript:alert(1)` in an `<a href>` can execute arbitrary JavaScript. The scheme validation rejects non-HTTP/HTTPS schemes.
- **Logging full URLs with query parameters**: URL query strings often contain API keys, session tokens, PII, or other sensitive data. Always log a hashed or truncated version.

### Senior Tips

Even without authentication, treat every URL submission as untrusted input. The URL field could contain XSS payloads in `javascript:` scheme attempts — this is why scheme validation (http/https only) is the first line of defense. When adding authentication, use a middleware that validates JWT tokens before any handler logic executes. Keep the auth middleware separate from business logic (it's cross-cutting, like logging).

### References

- See [Validation Rules](#validation-rules) for input validation details
- See [Error Handling](#error-handling) for error information leakage prevention
- See [Future Improvements](#future-improvements) for authentication roadmap

---

## Performance

### Indexes for Query Performance

The single most impactful optimization is the partial unique index on `short_code`:

```sql
CREATE UNIQUE INDEX idx_urls_short_code ON urls(short_code) WHERE deleted_at IS NULL;
```

This B-tree index ensures O(log n) lookups for every redirect request and every API GET operation. Without it, every lookup would be a sequential scan of the entire table.

### Connection Pool Tuning

The `pgx` connection pool (configured via `database/sql`) should be tuned for the expected workload:

| Setting | Recommended | Rationale |
|---|---|---|
| `MaxOpenConns` | 25 | Prevents connection exhaustion under high concurrency |
| `MaxIdleConns` | 5 | Reduces connection establishment overhead for reused connections |
| `ConnMaxLifetime` | 5m | Prevents stale connections; PostgreSQL kills idle connections after 10m by default |
| `ConnMaxIdleTime` | 2m | Recycles idle connections more aggressively than lifetime |

### Redirect Performance

The redirect flow should complete in < 50ms p99. This requires:

1. **Single database round-trip** where possible (consider combining SELECT and UPDATE via `UPDATE ... RETURNING`)
2. **Proper indexing** (partial unique index on short_code)
3. **Connection pooling** (tuned pool size)
4. **No unnecessary serialization/deserialization** in the handler (avoid base64 encoding of URLs)

### Atomic Updates

Redirect count increments must be atomic. The single-query approach is preferred:

```sql
UPDATE urls SET redirect_count = redirect_count + 1, updated_at = NOW() WHERE short_code = $1 AND deleted_at IS NULL RETURNING original_url
```

This combines the existence check, count increment, and result retrieval in one database round-trip, eliminating the race condition between SELECT and UPDATE in a concurrent scenario.

### Caching Opportunities

Caching is deferred to a future improvement (see [Future Improvements](#future-improvements)). For the initial implementation, every redirect hits the database. This is acceptable for the MVP scale (< 10K req/s) and avoids the complexity of cache invalidation. When cache hits reach 95%+ of redirects (due to popular URL concentration), Redis integration becomes worthwhile.

### Benchmarks

| Operation | Target | Measurement |
|---|---|---|
| Redirect (GET /r/{code}) | < 50ms p99 | p99 latency under 100 concurrent users |
| Create URL (POST /api/v1/urls) | < 200ms p99 | p99 latency |
| Get URL Info (GET /api/v1/urls/{code}) | < 30ms p99 | p99 latency |
| DB connection establishment | < 10ms | Single warm connection |

### Checklist

- [ ] Partial unique index on short_code (WHERE deleted_at IS NULL)
- [ ] Index on short_code for all GET/PUT/DELETE operations
- [ ] Connection pool tuned (MaxOpenConns, MaxIdleConns, ConnMaxLifetime)
- [ ] Redirect flow optimized (UPDATE ... RETURNING or single round-trip)
- [ ] No N+1 query patterns
- [ ] No unnecessary serialization in hot path
- [ ] Caching deferred to future (Redis) with clear extension point
- [ ] Performance benchmarks documented with target values
- [ ] `EXPLAIN ANALYZE` results checked for all query patterns

---

## Scalability

### Vertical Scaling

Vertical scaling (larger server) is the simplest approach. Increase CPU, RAM, and IOPS. PostgreSQL handles vertical scaling well for up to ~100K req/s on appropriate hardware.

### Horizontal Scaling

The service is **stateless** — all URL data lives in PostgreSQL, not in application memory. This means horizontal scaling is straightforward: deploy multiple instances behind a load balancer. No session affinity or sticky sessions are needed.

### Database Scaling

For read-heavy workloads (URL shorteners are typically 100x reads to 1 write):

1. **Read replicas**: Route redirect (GET) traffic to PostgreSQL read replicas. Write operations (POST, PUT, DELETE) go to the primary.
2. **Connection pooling**: Use `PgBouncer` or `pgpool-II` to multiplex many application connections through a smaller pool of database connections.
3. **Partitioning**: For 100M+ URLs, partition the `urls` table by `created_at` range.

### Background Jobs (Future)

Redirect counter increments could be decoupled from the HTTP response using a message queue (see [Future Improvements](#future-improvements)). The redirect handler would return immediately after serving the Location header, and a background worker would update the counter asynchronously. This improves redirect latency at the cost of eventual consistency.

### Future Message Queue

RabbitMQ or Kafka can handle background processing tasks (analytics aggregation, URL cleanup, bulk operations). The pattern: the service publishes an event to the queue, and a worker consumes it. This keeps the HTTP request fast and moves heavy processing off the request path.

### Future Microservices

If the system grows beyond a single codebase, it can decompose into:

- **URL Service**: URL CRUD and redirect (this service)
- **Analytics Service**: Click tracking and reporting
- **User Service**: Authentication, account management, custom aliases
- **API Gateway**: Entry point routing, rate limiting, auth

Each service uses the Repository pattern internally, and communication happens via HTTP APIs or message queues.

### Checklist

- [ ] Service is stateless (no in-memory session state)
- [ ] Connection pool configured and tuned
- [ ] Read replicas supported (future architecture)
- [ ] Connection pooling with PgBouncer planned
- [ ] Table partitioning strategy documented for > 100M rows
- [ ] Background job pattern documented for future message queue integration
- [ ] Horizontal scaling approach documented (stateless design)


---

## Testing Strategy

### Testing Philosophy

Tests are not an afterthought — they are core to the development workflow. The testing pyramid guides the approach:

- **Many fast unit tests (70%)** — test individual functions in isolation with mocks
- **Some integration tests (20%)** — test real database interactions with Dockerized PostgreSQL
- **Few E2E tests (10%)** — test the full HTTP request/response cycle with `httptest`

### Unit Tests

Unit tests verify individual functions in isolation. Mock all external dependencies.

**What to unit test:**
- `RandomShortener.Generate()` — always produces 8-char valid base62 strings, no collisions in 10K+ iterations
- `ValidateURL()` — accepts valid URLs, rejects invalid (no scheme, no host, too long, empty, whitespace-only)
- Service methods with mock repositories — business logic handles repository errors, not-found cases, gone cases
- `URL` struct creation — default values (redirect_count=0, deleted_at=nil)

**What NOT to unit test:**
- Database connectivity (integration test)
- HTTP server startup (integration/E2E test)
- Trivial getter/setter code (no logic)

### Integration Tests

Integration tests verify components work together with real dependencies.

**What to integration test:**
- Repository methods against real PostgreSQL (Docker test container)
- Goose migrations applied in order
- Service layer using real repository (not mocks)
- HTTP handlers using real service layer via `httptest`

**Setup:** Use `testcontainers-go` or manual Docker to spin up a PostgreSQL container for tests. Each test suite creates its own schema, inserts test data, runs tests, and cleans up.

### Repository Tests

Focus on CRUD operations and concurrency safety.

| Test | Scenario |
|---|---|
| Create → GetByShortCode | URL is retrievable immediately after creation |
| GetByShortCode → not found | Returns ErrURLNotFound for nonexistent shortcode |
| GetByShortCode → excluded deleted | Soft-deleted URLs not returned by GetByID or GetByShortCode |
| Update → original URL changed | UpdatedAt changed, new URL stored |
| SoftDelete → deleted_at set | URL not returned by subsequent Get queries |
| IncrementRedirectCount → count incremented | Atomic UPDATE with RETURNING |
| Concurrent Create → uniqueness | Database UNIQUE constraint catches duplicates |
| Concurrent Redirect → count accurate | Atomic UPDATE prevents lost updates |

### Handler Tests

Use `httptest.NewRecorder()` to simulate HTTP requests without starting a real server.

| Handler | Key Tests |
|---|---|
| CreateURL handler | 201 on success, 400 on missing/invalid URL, 500 for internal errors |
| Redirect handler | 302 with Location header, 404 for not found, 410 for deleted |
| GetURL handler | 200 with body, 404 for not found |
| UpdateURL handler | 200 on success, 404 for not found |
| DeleteURL handler | 204 on success, 404 for not found |
| Stats handler | 200 with redirect_count, 404 for not found |
| All handlers | Consistent error response format (code, message, request_id) |

### Service Tests

Use hand-written mock repositories. Each mock method returns configurable outcomes.

```go
type MockURLRepository struct {
    CreateFunc         func(ctx context.Context, url *domain.URL) error
    GetByShortCodeFunc func(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
    GetAllFunc         func(ctx context.Context) ([]domain.URL, error)
    UpdateFunc         func(ctx context.Context, code domain.ShortCode, url *domain.URL) error
    SoftDeleteFunc     func(ctx context.Context, code domain.ShortCode) error
    IncrementCountFunc func(ctx context.Context, code domain.ShortCode) (*domain.URL, error)
}
```

### Mock Strategy

Hand-written mocks are preferred over mock generation tools (`mockery`, `go mockgen`). They are easier to understand, maintain, and debug in a Go codebase.

**Mock behavior table:**

| Mock Field | Success Behavior | Error Behavior |
|---|---|---|
| `CreateFunc` | Returns nil error | Returns specific repository error |
| `GetByShortCodeFunc` | Returns URL entity | Returns ErrURLNotFound |
| `UpdateFunc` | Returns updated URL | Returns ErrURLNotFound |
| `SoftDeleteFunc` | Returns nil error | Returns ErrURLNotFound |
| `IncrementCountFunc` | Returns updated URL | Returns ErrURLNotFound |

### Coverage Goal

**80% code coverage** minimum. Measured with `go test ./... -cover`. Coverage should be reported per package, not just project-wide.

**Coverage targets by package:**

| Package | Target |
|---|---|
| `domain/` | 100% |
| `validator/` | 100% |
| `shortener/` | 100% |
| `service/` | 90%+ |
| `repository/postgres/` | 90%+ |
| `handler/` | 85%+ |
| `middleware/` | 80%+ |
| `error/` | 100% |
| `health/` | 100% |
| `config/` | 90%+ |
| `logger/` | 90%+ |

### Testing Milestones

| Milestone | Verification |
|---|---|
| Domain tests pass | `go test ./internal/domain/...` |
| Storage + Repo tests pass | `go test ./internal/storage/... ./internal/repository/postgres/...` |
| Shortener tests pass | `go test ./internal/shortener/...` |
| Validator tests pass | `go test ./internal/validator/...` |
| Service tests pass | `go test ./internal/service/...` |
| Handler tests pass | `go test ./internal/handler/...` |
| Middleware tests pass | `go test ./internal/middleware/...` |
| Full suite with race detector | `go test -race -cover ./...` |
| Coverage > 80% | `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` |

### Testing Checklist (Per Package)

**`domain/`:** [ ] URL struct with all fields, [ ] IsValidStatus for valid/invalid values, [ ] Sentinel errors non-nil and comparable, [ ] JSON tags correct, [ ] DeletedAt nil by default

**`shortener/`:** [ ] Generate returns 8 chars, [ ] Valid base62 only, [ ] No collisions in 10K iterations, [ ] Unique across calls

**`validator/`:** [ ] Accepts valid URLs (http, https, with host), [ ] Rejects URL without scheme, [ ] Rejects URL without host, [ ] Rejects URL > 2048 chars, [ ] Rejects empty string, [ ] Rejects whitespace-only, [ ] ValidateShortCode 8 chars base62, [ ] ValidateID accepts positive IDs, [ ] Boundary testing (0, 1, 8, 2047, 2048, 2049)

**`service/`:** [ ] CreateURL valid → success with StatusTodo, [ ] CreateURL empty URL → error (no I/O), [ ] CreateURL invalid scheme → error, [ ] CreateURL > 2048 chars → error, [ ] CreateURL repo error → propagated, [ ] GetURL existing → success, not found → error, deleted → gone, [ ] ListURLs empty → empty slice, all → all active, [ ] UpdateURL existing → updated, not found → error, [ ] DeleteURL existing → success, not found → error, [ ] Redirect increment → count increments, not found → error, [ ] Redirect deleted → gone error, [ ] All methods handle all repository error cases

**`repository/postgres/`:** [ ] Create returns ID and timestamps, [ ] GetByShortCode returns URL, [ ] GetByShortCode not found → error, [ ] GetByShortCode excludes deleted, [ ] Update modifies URL and UpdatedAt, [ ] SoftDelete sets deleted_at (idempotent), [ ] IncrementRedirectCount increases counter, [ ] Unique constraint violation caught, [ ] All queries parameterized (no SQL injection), [ ] Use QueryRowContext not QueryRow

**`handler/`:** [ ] CreateURL returns 201 with body, [ ] CreateURL missing body → 400, [ ] CreateURL invalid URL → 400, [ ] CreateURL internal error → 500 (masked), [ ] GetURL returns 200 with body, [ ] GetURL not found → 404, [ ] DeleteURL success → 204, [ ] DeleteURL not found → 404, [ ] Redirect returns 302 with Location, [ ] Redirect not found → 404, [ ] Redirect deleted → 410, [ ] All error responses follow consistent {error: {code, message, request_id}} format, [ ] request_id matches log correlation

**`middleware/`:** [ ] Request ID added to every request, [ ] Request ID propagated to context, [ ] Logging middleware logs method/path/status/duration, [ ] Recovery middleware catches panics and returns 500, [ ] Recovery middleware logs panic details

**`health/`:** [ ] Returns 200 when DB connected, [ ] Returns 503 when DB disconnected

### Senior Tips

Write tests first (TDD) for critical paths. Using `t.Parallel()` in unit tests significantly reduces test suite runtime. Use `t.Helper()` to mark helper functions so they are excluded from line coverage counting.

### Common Mistakes

- **Testing implementation details instead of behavior**: Test what the function does, not how it does it. If you refactor the internal algorithm, the test should still pass.
- **Shared mutable state between tests**: Each test should create its own fake data and clean up after itself. Do not rely on test execution order.
- **Ignoring error testing**: Every happy path should have corresponding error path tests. A test suite with only happy paths gives false confidence.
- **Using production database for tests**: Always use isolated test databases (Docker containers or temp files). Never run tests against shared dev or production databases.

### Checklist

- [ ] Testing philosophy documented (testing pyramid)
- [ ] Unit testing strategy defined with mock repository approach
- [ ] Integration testing strategy defined (Dockerized PostgreSQL)
- [ ] Repository test matrix defined
- [ ] Handler test matrix defined
- [ ] Service test matrix defined with mock repository
- [ ] Hand-written mock repository pattern documented
- [ ] Coverage goal set (80% overall, 90%+ per core package)
- [ ] Per-package test checklist provided
- [ ] Test milestones table defined
- [ ] `go test -race` documented as requirement
- [ ] `go test -cover` documented as requirement
- [ ] Testing strategy documented for all layers (domain, storage, repository, service, handler, middleware)


---

## Development Roadmap

The implementation is broken into progressive phases, building from the top (cmd entry point) down to the database. Follow them in order. Each phase builds on the previous one.

---

### Phase 1 — Project Bootstrap

**Goal:** Working Go project that compiles with zero dependencies. The build pipeline works.

**Tasks:**
- ✅ `go mod init github.com/mmk31585/url-shortener`
- [ ] Create directory structure (`cmd/server/`, all `internal/*/` packages)
- ✅ Create placeholder `.go` files in each package directory
- ✅ Create `cmd/server/main.go` with minimal `func main()` that prints help
- ✅ Create `Makefile` with targets: `build`, `test`, `fmt`, `vet`, `clean`
- ✅ Create `.gitignore` (Go artifacts, env files, coverage)
- ✅ Create `.env.example` with all environment variable placeholders
- ✅ Verify `go build ./...` succeeds
- ✅ Verify `go vet ./...` succeeds
- ✅ Verify `go test ./...` succeeds
- ✅ Verify `gofmt -s .` produces no diffs

**Deliverables:**
- ✅ Compilable project
- ✅ Directory structure exists for all packages
- ✅ `go build`, `go vet`, `go test` all pass
- ✅ Makefile works

**Suggested Commit Messages:**
```
chore: bootstrap Go module and project skeleton
chore: add Makefile with build, test, fmt, vet targets
chore: add .gitignore and .env.example
```
it needs to wire everything: load config → create logger → connect to DB → run migrations → start server.

**Definition of Done:**
- [ ] `go build -o /dev/null ./cmd/server` produces a binary
- [ ] `go vet ./...` is clean
- [ ] Every package directory has at least one `.go` file
- [ ] `go.mod` exists with correct module path
- [ ] Makefile has working `build`, `test`, `fmt`, `vet` targets

---

### Phase 2 — Config & Logger

**Goal:** Configuration loading and structured logging — the two cross-cutting foundations every other layer needs.

**Tasks:**
- [ ] `internal/config/config.go` — `Config` struct with `Load()` function
- [ ] `internal/logger/logger.go` — `New(env string) *slog.Logger`
- [ ] Config reads env vars via `os.LookupEnv` with sensible defaults
- [ ] Logger uses text handler for development, JSON handler for production
- [ ] Wire config + logger in `main.go`
- [ ] Verify `go build ./...` succeeds
- [ ] Verify `go vet ./...` succeeds

**Deliverables:**
- [ ] Config loads from env vars with all fields documented
- [ ] Logger works in both development and production modes
- [ ] `main.go` wires config and logger
- [ ] All config fields have correct types

**Suggested Commit Messages:**
```
feat: add configuration loading from environment variables
feat: add slog structured logging with dev/prod modes
feat: wire config and logger in main.go
```

**Definition of Done:**
- [ ] Config loads all environment variables with correct types
- [ ] Logger outputs JSON in production mode, text in development mode
- [ ] Configuration validation fails early (missing required vars)
- [ ] No hardcoded configuration values in business logic

---

### Phase 3 — Domain Layer

**Goal:** Core domain types — `URL` entity, `ShortCode` type, status, and domain errors. Zero external dependencies in this package.

**Tasks:**
- [ ] `internal/domain/url.go` — `URL` struct, `ShortCode` type, `Status` type with constants
- [ ] `internal/domain/errors.go` — Sentinel errors (`ErrURLNotFound`, `ErrURLAlreadyDeleted`, etc.)
- [ ] All fields have correct types and JSON tags
- [ ] `DeletedAt` is `*time.Time` (pointer for nil check)
- [ ] No methods on the URL struct (pure data container)
- [ ] All errors are package-level `var` declarations (comparable with `errors.Is`)
- [ ] Package imports only `time`, `encoding/json`, and `errors` from stdlib
- [ ] Verify `go build ./internal/domain/...` succeeds
- [ ] Verify `go vet ./internal/domain/...` is clean

**Deliverables:**
- [ ] URL entity with all 7 fields defined
- [ ] ShortCode type defined (typed string)
- [ ] Status type with constants
- [ ] Sentinel errors defined and documented
- [ ] Zero external dependencies in domain package

**Suggested Commit Messages:**
```
feat(domain): add URL entity with ShortCode, Status, and timestamps
feat(domain): add domain sentinel errors
```

**Definition of Done:**
- [ ] All domain types compile and pass tests
- [ ] Package has zero imports from other project packages
- [ ] All JSON tags are snake_case
- [ ] `DeletedAt` is nil by default (not `*time.Time` pointing to zero time)
- [ ] Coverage > 95%

---

### Phase 4 — Storage (PostgreSQL + Goose)

**Goal:** Database connection, Goose migrations, and storage layer ready.

**Tasks:**
- [ ] Add Goose (`github.com/pressly/goose/v3`) andpgx (`github.com/jackc/pgx/v5`) dependencies
- [ ] Create `migrations/` directory
- [ ] Write `000001_create_urls_table.up.sql` (id, short_code, original_url, redirect_count, created_at, updated_at, deleted_at)
- [ ] Write `000001_create_urls_table.down.sql` (DROP TABLE)
- [ ] Write `000002_add_indexes.up.sql` (unique partial index, regular indexes)
- [ ] Write `000002_add_indexes.down.sql` (DROP INDEX)
- [ ] `internal/storage/storage.go` — `Storage` struct with `*sql.DB` field
- [ ] `New(url string) (*Storage, error)`: opens DB, pings, auto-runs Goose `goose.Up`, returns Storage
- [ ] `(*Storage) DB() *sql.DB` method (gives repository access)
- [ ] `(*Storage) Close() error` method
- [ ] Auto-run migrations on startup (or via CLI flag)

**Deliverables:**
- [ ] Goose working with numbered migration files
- [ ] urls table created with all columns, constraints, indexes
- [ ] Storage struct provides DB access
- [ ] Migrations apply automatically on first connection

**Suggested Commit Messages:**
```
feat(storage): add PostgreSQL connection with Goose auto-migrations
feat(db): add initial migration for urls table with constraints
feat(db): add indexes and partial unique index on short_code
```

**Definition of Done:**
- [ ] All migrations apply cleanly (`goose up`)
- [ ] All down migrations work (`goose down`)
- [ ] `New()` connects to PostgreSQL and applies pending migrations
- [ ] `Storage.DB()` returns a valid `*sql.DB`
- [ ] `Storage.Close()` closes the connection properly

---

### Phase 5 — Repository

**Goal:** `URLRepository` interface + PostgreSQL implementation.

**Tasks:**
- [ ] `internal/repository/repository.go` — Define `URLRepository` interface with 7 methods: `Create`, `GetByShortCode`, `GetAll`, `GetByOriginalURL`, `Update`, `SoftDelete`, `IncrementRedirectCount`
- [ ] `internal/repository/postgres/url.go` — `PostgresURLRepository` struct implementing `URLRepository`
- [ ] `NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository` constructor
- [ ] Implement all 7 methods using parameterized queries
- [ ] Every query uses `QueryRowContext` or `QueryContext`
- [ ] Every scan uses proper types
- [ ] Database errors mapped to domain errors at repository boundary
- [ ] Partial index on short_code (where deleted_at IS NULL) enables shortcode reuse

**Deliverables:**
- [ ] `*PostgresURLRepository` satisfies `URLRepository` (compiler check)
- [ ] All CRUD operations work correctly
- [ ] Soft delete excludes deleted URLs from all read operations
- [ ] Redirect count increments atomically
- [ ] Unique constraint violation handled gracefully

**Suggested Commit Messages:**
```
feat(repository): add URLRepository interface
feat(repository): add PostgresURLRepository with all CRUD operations
```

**Definition of Done:**
- [ ] All 7 repository methods implemented
- [ ] All queries use parameterized statements (no SQL injection)
- [ ] All queries use context-aware methods
- [ ] Database errors mapped to domain errors before leaving repository
- [ ] Soft delete correctly handled
- [ ] All integration tests pass against real PostgreSQL
- [ ] Coverage > 90%

---

### Phase 6 — Shortcode Generator (Strategy)

**Goal:** `Shortener` interface + `RandomShortener` implementation using crypto/rand + base62.

**Tasks:**
- [ ] `internal/shortener/shortener.go` — Define `Shortener` interface with `Generate() (domain.ShortCode, error)`
- [ ] `RandomShortener` struct with `length int` (default 8) and `alphabet string` (base62)
- [ ] Implement `Generate()` using `crypto/rand.Read()` + modulo mapping to alphabet
- [ ] Validate generated shortcode is exactly 8 chars and valid base62

**Deliverables:**
- [ ] Strategy interface allows swapping algorithms
- [ ] RandomShortener produces unique 8-char base62 shortcodes
- [ ] No collisions in 10K+ iterations

**Suggested Commit Messages:**
```
feat(shortener): add Shortener interface and RandomShortener strategy
```

**Definition of Done:**
- [ ] `Generate()` always returns exactly 8 characters
- [ ] Generated codes contain only valid base62 characters
- [ ] No collisions across 10K+ random generations
- [ ] Coverage 100%

---

### Phase 7 — Validator

**Goal:** Pure validation functions with no side effects.

**Tasks:**
- [ ] `internal/validator/validator.go` — `ValidateURL(url string) error`
  - [ ] Empty/whitespace → error
  - [ ] `url.Parse` fails → error
  - [ ] Scheme not http/https → error
  - [ ] Host empty → error
  - [ ] Length > 2048 → error
- [ ] `ValidateShortCode(code string) error` — exactly 8 chars, valid base62
- [ ] `ValidateID(id int64) error` — id > 0

**Deliverables:**
- [ ] All validation functions are pure (no I/O, no state)
- [ ] Boundary conditions tested (0, 1, 7, 8, 9, 2047, 2048, 2049)
- [ ] Coverage 100%

**Suggested Commit Messages:**
```
feat(validator): add URL, shortcode, and ID validation functions
```

**Definition of Done:**
- [ ] All validation functions are pure (no side effects)
- [ ] All boundary conditions tested
- [ ] Coverage 100%
- [ ] `gofmt` passes

---

### Phase 8 — Service Layer

**Goal:** `URLService` interface + implementation with all business logic.

**Tasks:**
- [ ] `internal/service/service.go` — Define `URLService` interface (7 methods)
- [ ] `internal/service/url_service.go` — `urlService` struct implementing all methods
- [ ] `NewURLService(repo repository.URLRepository, shortener shortener.Shortener, maxURLLength int) *urlService`
- [ ] Implement `CreateURL`: validate → generate shortcode → persist → return URL
- [ ] Implement `GetURL`: fetch by shortcode, return or error
- [ ] Implement `ListURLs`: return all active URLs
- [ ] Implement `UpdateURL`: validate → find → modify → persist → return URL
- [ ] Implement `DeleteURL`: find → soft delete
- [ ] Implement `Redirect`: find → increment count → return original URL
- [ ] Implement `GetStats`: fetch URL with redirect count
- [ ] All business rules (BR-01 to BR-10) enforced at service layer
- [ ] Service has zero HTTP or database imports

**Deliverables:**
- [ ] All 7 use-case methods work correctly with mock repository
- [ ] All validation delegated to validator package
- [ ] All shortcode generation delegated to shortener package
- [ ] All data access delegated to repository interface
- [ ] Service contains zero business logic leaks into HTTP or DB layers
- [ ] Coverage > 90%

**Suggested Commit Messages:**
```
feat(service): add URLService interface
feat(service): implement CreateURL, GetURL, ListURLs
feat(service): implement UpdateURL, DeleteURL, Redirect, GetStats
```

**Definition of Done:**
- [ ] All 7 service methods work correctly with mock repositories
- [ ] All error paths tested (not found, gone, validation, etc.)
- [ ] Business rules BR-01 to BR-10 validated by tests
- [ ] Coverage > 90% on service package

---

### Phase 9 — Handler Layer (HTTP Handlers)

**Goal:** All HTTP endpoints working. Handlers are thin — they parse, delegate, format.

**Tasks:**
- [ ] `internal/handler/handler.go` — BaseHandler struct with service + logger
- [ ] `internal/handler/url.go` — All 7 endpoint handlers
  - [ ] `CreateURL` → POST /api/v1/urls → 201 on success, 400 on validation error
  - [ ] `GetURL` → GET /api/v1/urls/{shortcode} → 200 or 404
  - [ ] `Redirect` → GET /r/{shortcode} → 302 with Location header (302, not 301!)
  - [ ] `UpdateURL` → PUT /api/v1/urls/{shortcode} → 200 on success
  - [ ] `DeleteURL` → DELETE /api/v1/urls/{shortcode} → 204 on success
  - [ ] `GetStats` → GET /api/v1/urls/{shortcode}/stats → 200 or 404
  - [ ] `Health` → GET /health → 200 with DB status
- [ ] Each handler returns consistent error format `{error: {code, message, request_id}}`
- [ ] Handlers contain zero business logic, zero SQL, zero net/url calls

**Deliverables:**
- [ ] All 7 endpoints respond correctly for success and error cases
- [ ] 302 redirect (not 301)
- [ ] Error responses are user-friendly with request_id correlation
- [ ] Handler tests using `httptest` for all endpoints

**Suggested Commit Messages:**
```
feat(handler): add all HTTP handlers for URL CRUD, redirect, and stats
```

**Definition of Done:**
- [ ] All 7 endpoints work correctly via curl
- [ ] Handlers contain zero business logic
- [ ] Error responses follow consistent format
- [ ] Coverage > 85%

---

### Phase 10 — Router + Middleware

**Goal:** Route definitions and cross-cutting middleware wired up.

**Tasks:**
- [ ] `internal/router/router.go` — `NewRouter(svc, logger) *http.ServeMux` with all 8 routes
- [ ] `internal/middleware/middleware.go` — Three middleware functions:
  - [ ] **Request ID middleware**: Generate UUID v4, store in context, add `X-Request-ID` header to response
  - [ ] **Logging middleware**: Log method, path, status, duration, request_id for every request
  - [ ] **Recovery middleware**: Catch panics, log details, return 500 generic message
- [ ] Middleware applied in correct order (outer-to-inner)
- [ ] Wire middleware in `main.go`

**Deliverables:**
- [ ] All routes bound correctly
- [ ] Request ID generated per request, propagated through context and responses
- [ ] Every request logged with metadata
- [ ] Panics caught and returned as clean 500 responses

**Suggested Commit Messages:**
```
feat(middleware): add request ID, logging, and recovery middleware
feat(router): set up HTTP routes with ServeMux
```

**Definition of Done:**
- [ ] All 8 routes registered
- [ ] All 3 middleware execute for every request
- [ ] Logging includes all metadata fields
- [ ] Panics are caught without crashing the server

---

### Phase 11 — Error Handling Polish

**Goal:** Consistent error wrapping, user-friendly messages, and panic recovery at the top level.

**Tasks:**
- [ ] `internal/error/error.go` — `AppError` struct with Code, Message, StatusCode
- [ ] Constructors: `NewValidationError`, `NewNotFoundError`, `NewConflictError`, `NewGoneError`, `NewInternalError`
- [ ] Ensure all errors use `%w` wrapping at every layer boundary
- [ ] In handlers, use `errors.Is` to detect sentinel errors
- [ ] Add top-level `defer/recover` in `main.go` to catch unhandled panics
- [ ] All `fmt.Fprintln(os.Stderr, ...)` use consistent format

**Deliverables:**
- [ ] Error chain preserved across all layers (repository → service → handler → HTTP response)
- [ ] No raw database errors reach HTTP clients
- [ ] Panics are caught at the top level and return clean 500 responses
- [ ] All error messages are user-friendly and include request_id

**Suggested Commit Messages:**
```
refactor(errors): add AppError type and consistent error propagation
```

**Definition of Done:**
- [ ] `errors.Is` works across all layers for each sentinel error
- [ ] No raw errors printed to users
- [ ] All error responses include request_id
- [ ] No panics leak to the client

---

### Phase 12 — Swagger / OpenAPI Documentation

**Goal:** Interactive API documentation.

**Tasks:**
- [ ] Create `docs/swagger.yaml` with OpenAPI 3.0 specification
- [ ] Document all 8 endpoints with request/response schemas
- [ ] Include example request and response bodies for every endpoint
- [ ] Define shared components for `URL`, `ErrorResponse`, `StatsResponse`
- [ ] Tag endpoints as `/urls` and `/redirect`

**Deliverables:**
- [ ] Swagger YAML is valid and parseable
- [ ] All endpoints documented with examples at every status code
- [ ] Error responses documented for each endpoint

**Suggested Commit Messages:**
```
docs: add Swagger/OpenAPI API documentation
```

**Definition of Done:**
- [ ] All 8 endpoints documented
- [ ] YAML is valid
- [ ] Request/response schemas match actual implementations

---

### Phase 13 — Docker & Deployment

**Goal:** Containerized application ready for deployment.

**Tasks:**
- [ ] Multi-stage `Dockerfile` (golang:1.22-alpine builder → alpine:3.19 runtime)
- [ ] `CGO_ENABLED=0`, `ldflags="-s -w"` for small static binary
- [ ] `docker-compose.yml` with App + PostgreSQL services
- [ ] `.dockerignore` excluding build artifacts, tests, docs
- [ ] `HEALTHCHECK` in Dockerfile
- [ ] Persistent volume for PostgreSQL data
- [ ] Environment variables set via env vars (not hardcoded)

**Deliverables:**
- [ ] Docker image builds successfully
- [ ] Image is small (< 50MB) and contains only binary + CA certs
- [ ] `docker-compose up` starts the full stack
- [ ] Application connects to PostgreSQL through Docker network
- [ ] Health check works

**Suggested Commit Messages:**
```
feat(docker): add multi-stage Dockerfile and docker-compose for local dev
```

**Definition of Done:**
- [ ] `docker build -t url-shortener .` succeeds
- [ ] `docker-compose up` starts both services
- [ ] Health endpoint responds correctly inside Docker
- [ ] Data persists across container restarts

---

### Phase 14 — Code Quality & Refactoring

**Goal:** Review and polish the entire codebase.

**Tasks:**
- [ ] `gofmt -s .` produces no diffs
- [ ] `goimports` produces no diffs
- [ ] `go vet ./...` is clean
- [ ] `go mod tidy` has been run
- [ ] SOLID principles verified (see [Section 13](#soli-principles))
- [ ] No package has more than one reason to change
- [ ] No function exceeds 30 lines
- [ ] No magic numbers or strings in business logic
- [ ] No commented-out code
- [ ] No unused imports or variables
- [ ] No `init()` functions (except `sync.Once`)
- [ ] No global variables (except sentinel errors)
- [ ] No circular imports
- [ ] Go doc comments on all exported symbols
- [ ] Package-level doc comments on every package
- [ ] Import statements grouped correctly (stdlib → external → internal)
- [ ] DRY — no duplicate code across packages
- [ ] YAGNI — no speculative features

**Suggested Commit Messages:**
```
refactor: code quality review and cleanup
```

**Definition of Done:**
- [ ] Every quality check passes
- [ ] No warnings from fmt, vet, or golangci-lint
- [ ] Codebase follows project coding standards

---

### Phase 15 — Final Testing & Documentation

**Goal:** All tests passing, documentation complete and accurate.

**Tasks:**
- [ ] `go test -race -cover ./...` passes with > 80% coverage
- [ ] Every exported function has at least one test
- [ ] All error paths tested for every handler and service method
- [ ] Business rules BR-01 to BR-10 validated by tests
- [ ] Review README.md for accuracy against actual implementation
- [ ] Review Software Design Document for accuracy against code
- [ ] Ensure every ADR matches actual implementation decisions
- [ ] Update Architecture diagram if implementation differs from design
- [ ] Remove all TODO or placeholder comments
- [ ] Verify `go doc ./...` produces readable output

**Suggested Commit Messages:**
```
test: achieve 80%+ coverage with race detector
docs: finalize README and SDD for accuracy
test: add integration tests for all business rules
```

**Definition of Done:**
- [ ] `go test -race -cover ./...` passes
- [ ] README accurately reflects the implemented project
- [ ] All doc comments are complete
- [ ] No TODO/FIXME comments without tracking issues

---

### Phase 16 — Release v1.0.0

**Goal:** Tag and publish the first stable release.

**Tasks:**
- [ ] All previous phases complete and committed
- [ ] `main` branch is up to date
- [ ] `go build -o url-shortener ./cmd/server` produces a working binary
- [ ] Final manual smoke test of all endpoints (create, redirect, update, delete, stats, health)
- [ ] `make all` passes (fmt → vet → build → test)
- [ ] Docker build and compose work correctly
- [ ] Create and push tag: `git tag -a v1.0.0 -m "v1.0.0: Initial stable release"`

**Suggested Commit Messages:**
```
chore: bump version to v1.0.0
```

**Definition of Done (Project):**
- [ ] All functional requirements implemented
- [ ] All non-functional requirements satisfied
- [ ] Zero external dependencies beyond pgx and goose
- [ ] Binary compiled as a single static executable
- [ ] Test coverage exceeds 80%
- [ ] All code quality checks pass
- [ ] Documentation is complete and accurate
- [ ] Git tag `v1.0.0` published
- [ ] Project is ready for public consumption

---

## Git Strategy

### Branch Naming Convention

| Branch Type | Name Pattern | Purpose |
|---|---|---|
| Main | `main` | Production-ready, always compiles |
| Development | `develop` | Integration branch for ongoing work |
| Feature | `phase/<number>-<name>` | `phase/1-bootstrap`, `phase/2-domain`, etc. |
| Hotfix | `hotfix/<description>` | Urgent bug fixes |
| Release | `release/v1.0.0` | Release preparation |

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/): `<type>(<scope>): <subject>`

| Type | Usage |
|---|---|
| `feat` | New feature or endpoint |
| `fix` | Bug fix |
| `test` | Adding or updating tests |
| `docs` | Documentation changes |
| `refactor` | Code restructuring, no behavior change |
| `chore` | Maintenance (dependencies, tooling) |
| `ci` | CI/CD configuration changes |
| `config` | Configuration changes |
| `perf` | Performance improvement |
| `security` | Security-related changes |

### Pull Request Process

1. Branch from `main`, name it `phase/<number>-<name>` or `feat/<description>`
2. Keep PRs small and focused (one phase per PR)
3. Include checklist of items completed in the PR description
4. Run `make all` (fmt → vet → build → test) before opening PR
5. At least one approving review required
6. Squash merge to keep history clean

### Release Strategy

Use Semantic Versioning (`MAJOR.MINOR.PATCH`):
- `v1.0.0` — Initial stable release
- `v1.1.0` — Backward-compatible new features
- `v1.1.1` — Backward-compatible bug fixes
- `v2.0.0` — Breaking API changes (e.g., endpoint path changes)

### Checklist

- [ ] Branch naming convention defined
- [ ] Commit convention defined (Conventional Commits)
- [ ] PR process defined (review, CI checks, merge)
- [ ] Versioning strategy defined (SemVer)
- [ ] Branch protection rules for `main`
- [ ] `.gitignore` includes Go artifacts, `.env`, coverage files

### Git Branch Strategy Diagram

```
main ●────●────●────●──────────────────────────────────● v1.0.0
       \    / \    / \    /                             /
phase/1   ●──●   ●──●   ●──●                         /
phase/2       ●──●                             /
phase/3               ●──●                    /
phase/4                   ●──●                /
phase/5                       ●──●            /
phase/6                           ●──●        /
phase/7                               ●──●    /
phase/8                                   ●──●/
release/v1.0.0  ──────────────────────────●
```

**Rules:**
- `main` is always in a working state
- Each phase gets its own branch from `main`
- Merge phase branches back to `main` via PR after passing all tests
- After all phases, create a release branch for final testing and tagging

### Checklist

- [ ] Branch naming convention documented
- [ ] Commit convention (Conventional Commits) documented
- [ ] PR process defined
- [ ] Versioning strategy (SemVer) documented
- [ ] Branch protection rules for main defined
- [ ] Main branch always in working state

---

## ADR — Architecture Decision Records

### ADR-001: PostgreSQL

**Status:** Accepted
**Context:** The project needs persistent storage for URL mappings.
**Decision:** Use PostgreSQL for relational integrity, JSON support, and mature Go drivers.
**Alternatives:** MySQL (weaker standards compliance), SQLite (no concurrent writes), MongoDB (overkill for single entity), In-memory (no persistence).
**Trade-offs:** Requires separate DB process. Benefits outweigh the operational cost for a production-grade service.

### ADR-002: net/http (Standard Library)

**Status:** Accepted
**Context:** Go has many HTTP frameworks (Gin, Echo, Fiber, Chi). The project uses `net/http`.
**Decision:** Use Go's standard library HTTP server.
**Rationale:** Production-ready, zero dependencies, excellent performance, demonstrates Go idioms. For 6 endpoints, a framework adds unnecessary complexity.
**Alternatives:** Gin (external dependency, framework-specific patterns), Echo (similar), Chi (future option for complex routing).
**Trade-offs:** ServeMux lacks path parameters; project handles routing with path parsing. More verbose but more educational.

### ADR-003: log/slog (Standard Library Logger)

**Status:** Accepted
**Context:** Go has zap, zerolog, logrus. The project uses `log/slog`.
**Decision:** Use Go's standard structured logger.
**Rationale:** Built-in, zero dependencies, JSON/text output, log levels. For an educational project emphasizing stdlib, slog is correct.
**Alternatives:** zap (faster but external dep), zerolog (similar), logrus (deprecated).
**Trade-offs:** Slightly slower than zap for high-throughput logging. For the URL shortener's expected scale, slog is sufficient. Can adopt zap later if needed.

### ADR-004: Goose for Migrations

**Status:** Accepted
**Context:** Need database migration management. Options: Goose, golang-migrate, dbmate, manual scripts.
**Decision:** Use Goose with numbered SQL files.
**Rationale:** Mature, uses explicit SQL files, supports up/down migrations, good PostgreSQL integration.
**Alternatives:** golang-migrate (more complex), dbmate (less adoption), manual scripts (no version tracking).
**Trade-offs:** Adds a Go dependency but provides automated, versioned, reversible migrations.

### ADR-005: Repository Pattern (No ORM)

**Status:** Accepted
**Context:** Separate data access from business logic. Options: Repository Pattern, Active Record, Data Mapper with ORM.
**Decision:** Repository Pattern with hand-written SQL and `database/sql`.
**Rationale:** Clean separation. Mock repositories for testing. Swap storage backends without touching business logic. No ORM magic.
**Alternatives:** Active Record (violates SRP), GORM (implicit magic, harder to debug).
**Trade-offs:** Hand-written SQL is more verbose than an ORM but more explicit, testable, and transparent.

### ADR-006: Layered Architecture (Not Clean Architecture)

**Status:** Accepted
**Context:** Need an architectural style. Options: Layered, Clean Architecture (Hexagonal), Modular Monolith.
**Decision:** Layered Architecture (Presentation → Service → Repository → Storage).
**Rationale:** Simplest architecture providing separation of concerns. For 6 endpoints and 1 storage backend, Clean Architecture adds unnecessary layers.
**Alternatives:** Clean Architecture (over-engineered for this scope), Modular Monolith (warranted only at larger scale).
**Trade-offs:** Can become anemic if undisciplined. Project counteracts by keeping models pure and all behavior in services.

### ADR-007: SERIAL Primary Key

**Status:** Accepted
**Context:** URL ID needs unique auto-incrementing identifier. Options: SERIAL (32-bit), BIGSERIAL (64-bit), UUID, ULID.
**Decision:** SERIAL (32-bit auto-increment integer).
**Rationale:** Compact (4 bytes), monotonically increasing, sufficient for millions of rows. Internal only — clients use the shortcode.
**Alternatives:** BIGSERIAL (unnecessary overhead at expected scale), UUID (16 bytes, less readable, unordered), ULID (similar to UUID).
**Trade-offs:** SERIAL IDs are predictable but the shortcode provides actual protection. Migration to BIGSERIAL is straightforward if needed.

### Checklist

- [ ] ADR-001: PostgreSQL documented
- [ ] ADR-002: net/http documented
- [ ] ADR-003: slog documented
- [ ] ADR-004: Goose documented
- [ ] ADR-005: Repository Pattern (no ORM) documented
- [ ] ADR-006: Layered Architecture documented
- [ ] ADR-007: SERIAL primary key documented

---

## Deployment

### Docker Strategy

Multi-stage build for a small production image:

**Stage 1 — Builder:** `golang:1.22-alpine` — downloads dependencies, compiles static binary with `CGO_ENABLED=0` and `ldflags="-s -w"`.

**Stage 2 — Runtime:** `alpine:3.19` — copies only the compiled binary and CA certificates. No source code, no Go toolchain in the final image.

**Docker Compose:** Spins up both the application and PostgreSQL for local development. Includes persistent volume for data, environment variable configuration, and health check.

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Environment (development, production, test) |
| `SERVER_ADDRESS` | `:8080` | TCP address to listen on |
| `DATABASE_URL` | `postgres://urlshortener:secret@localhost:5432/urlshortener?sslmode=disable` | PostgreSQL connection string |
| `LOG_LEVEL` | `info` | Minimum log level (debug, info, warn, error) |
| `SHORTCODE_LENGTH` | `8` | Length of auto-generated shortcodes |
| `MAX_URL_LENGTH` | `2048` | Maximum allowed URL length |

### Health Check

The `/health` endpoint verifies database connectivity (`SELECT 1`) and returns the service status. Configure Docker `HEALTHCHECK` to poll this endpoint.

### Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals for graceful shutdown: stop accepting new connections, allow in-flight requests to complete, close the database connection pool.

### Checklist

- [ ] Dockerfile defined with multi-stage build
- [ ] Production binary is statically linked (CGO_ENABLED=0)
- [ ] Binary is stripped (ldflags="-s -w")
- [ ] Final image does not contain source code or Go toolchain
- [ ] docker-compose.yml for local development defined
- [ ] PostgreSQL service defined with persistent volume
- [ ] All environment variables documented
- [ ] HEALTHCHECK included in Dockerfile
- [ ] Graceful shutdown implemented
- [ ] Configuration loaded from environment variables
- [ ] No hardcoded configuration values in business logic

### References

- See [Technology Decisions](#technology-decisions) for Docker rationale
- See [Development Roadmap](#development-roadmap) for Phase 13 (Docker)
- See [Configuration](#) for all environment variables

---

## Future Improvements

These features are out of scope for the initial release but planned for future iterations:

### Redis Cache
Add Redis as a read-through cache for redirect lookups. Store shortcode → URL mappings with a TTL. On redirect: check Redis first, on miss query PostgreSQL and populate cache. Significantly reduces database load under high traffic.

### Custom Alias
Allow users to choose their own shortcode (e.g., `/r/my-brand`). Requires namespace conflict handling and longer validation.

### URL Expiration / TTL
Add `expires_at` column with automatic cleanup. Useful for campaign links that should stop working after a date.

### Authentication & Authorization
JWT or OAuth2-based auth to protect URL management endpoints. Role-based access control for admin features.

### Rate Limiting
Limit create/redirect requests per IP or API key using a token bucket algorithm. Protect against abuse.

### QR Code Generation
Generate QR codes for short URLs. Useful for business cards, printed materials, and product packaging.

### Analytics Dashboard
Track beyond redirect counts: referrers, user agents, geolocation, and timestamps. Requires a time-series data store and a frontend.

### Distributed ShortCode Generator
For multi-instance deployments, use Snowflake-like IDs or a distributed coordination service to avoid collisions across instances.

### Background Workers & Message Queue
Decouple redirect count updates from the HTTP response using RabbitMQ or Kafka. Improves redirect latency by making count increments asynchronous.

### CDN for Redirects
Configure a CDN edge to cache redirect responses. Reduces origin load and improves global redirect latency.

### Metrics & Monitoring
Prometheus metrics for request rates, latencies, error rates, and redirect counts. Grafana dashboards for visualization.

### Checklist

- [ ] Redis cache planned (future: read-through cache for redirects)
- [ ] Custom alias feature planned (future: user-chosen shortcodes)
- [ ] URL expiration / TTL planned (future: expires_at column)
- [ ] Authentication & Authorization planned (future: JWT-based auth)
- [ ] Rate limiting planned (future: token bucket per IP/API key)
- [ ] QR code generation planned (future: QR code image endpoint)
- [ ] Analytics dashboard planned (future: referrer, UA, geolocation tracking)
- [ ] Distributed ID generation planned (future: Snowflake-like for multi-instance)
- [ ] Background workers / message queue planned (future: decouple count updates)
- [ ] CDN integration planned (future: cache redirect responses at edge)
- [ ] Metrics & monitoring planned (future: Prometheus + Grafana)

### References

- See [Development Roadmap](#development-roadmap) for implementation order
- See [Performance](#performance) for caching opportunities section

---

## Learning Outcomes

Completing this project teaches the following backend engineering concepts:

### Go
- Structs, interfaces, and composition over inheritance
- Error handling (`error` as first-class value, wrapping with `%w`)
- `context.Context` for cancellation, deadlines, and request-scoped values
- JSON encoding/decoding for HTTP APIs
- `database/sql` package and connection pooling
- `net/http` standard library HTTP server
- `log/slog` structured logging

### REST & HTTP
- HTTP methods and semantics (GET, POST, PUT, DELETE)
- Status code semantics (302 vs 301, 4xx vs 5xx)
- Middleware pattern for cross-cutting concerns
- How a real HTTP server works from the ground up

### Clean Architecture & Design Patterns
- Layered architecture (Presentation → Service → Repository → Storage)
- Repository Pattern (interface-based data access)
- Strategy Pattern (shortcode generation swappable)
- Dependency Injection (manual, constructor-based, no framework)
- Interface Segregation (small, focused interfaces)
- Dependency Inversion (high-level modules don't depend on low-level modules, both depend on abstractions)

### SOLID Principles
- Single Responsibility (each layer/type has one clear reason to change)
- Open/Closed (repository and shortener interfaces extensible without modification)
- Liskov Substitution (any repository/shortener implementation can be swapped)
- Interface Segregation (focused interfaces, not fat ones)
- Dependency Inversion (reversed dependency arrows via interfaces)

### Database & SQL
- `CREATE TABLE` with constraints (PRIMARY KEY, UNIQUE, CHECK, NOT NULL)
- B-tree indexing on hot lookup paths
- Partial unique indexes (unique where deleted_at IS NULL)
- `TIMESTAMPTZ` and UTC storage
- Migration management with Goose
- Parameterized queries (SQL injection prevention)
- Soft delete patterns (`deleted_at` nullable column)

### Docker & Deployment
- Multi-stage builds for small production images
- Static binaries with `CGO_ENABLED=0`
- `docker-compose` for local development with dependent services
- `HEALTHCHECK` in Dockerfiles
- `.dockerignore` for build context optimization

### Swagger / OpenAPI
- Documenting REST APIs with hand-written OpenAPI 3.0 YAML
- Interactive API testing with Swagger UI
- Request/response examples in documentation

### Testing
- Unit testing in Go (`testing` package)
- Table-driven tests for parameterized test cases
- Hand-written mock repositories
- Integration testing with real PostgreSQL (Docker test container)
- Coverage measurement (`go test -cover`)
- Race detection (`go test -race`)
- Testing pyramid: many unit tests, some integration, few E2E

### Software Design & Engineering
- Clean Code naming conventions
- Small, focused functions (under 30 lines, ≤4 parameters)
- DRY principle (no duplicate code)
- KISS principle (prefer simplicity)
- YAGNI principle (no speculative features)
- Configuration externalization (Twelve-Factor App)
- Structured logging for observability
- ADRs for documenting architectural decisions
- Progressive implementation (bootstrap → domain → storage → service → handler)
- Definition of Ready / Definition of Done per phase

### Backend Engineering
- Dependency graph management (DAG, no circular imports)
- Domain-driven boundaries (domain has zero external dependencies)
- Cross-cutting concerns separated into middleware
- Error propagation across layer boundaries with `%w` wrapping
- `errors.Is` for sentinel error detection across wrapped error chains

### Trade-off Analysis
- Why 302 instead of 301 for redirects (301s are cached; prevent redirect count tracking)
- Why SERIAL instead of UUID (compact, sequential, internal)
- Why slog instead of zap (stdlib, zero deps; zap adopted later if needed)
- Why hand-written SQL instead of ORM (explicit, testable, transparent)
- Why layered architecture instead of Clean Architecture (simpler for this scope)
- When to stop adding features and ship

### Interview Knowledge & CV Skills
- This project demonstrates production-oriented Go backend engineering
- Clean Architecture, SOLID, DDD, repository pattern, and layered architecture in a real project
- Database design with constraints, indexes, and migrations
- Docker containerization and deployment
- Testing at multiple levels (unit, integration, E2E)

### Checklist

- [ ] Go: interfaces, error handling, context, JSON, database/sql
- [ ] HTTP: methods, status codes, middleware, ServeMux
- [ ] REST: resource design, versioning, consistent error formats
- [ ] Repository Pattern: interfaces, mocking, swapping backends
- [ ] Layered Architecture: separation of concerns, dependency direction
- [ ] SQL/PostgreSQL: tables, constraints, indexes, migrations, parameterized queries, soft delete
- [ ] Docker: multi-stage builds, compose, health checks, .dockerignore
- [ ] Swagger/OpenAPI: API documentation, interactive testing
- [ ] Testing: unit, integration, mocks, coverage, race detection, test pyramid
- [ ] Dependency Injection: constructor-based, manual, testability
- [ ] Design Principles: SOLID, Clean Code, patterns, trade-offs
- [ ] Architecture thinking: ADRs, layered vs hexagonal, stdlib vs frameworks
- [ ] Engineering trade-offs: documented with rationale

---

## Engineering Notes

### Key Lessons Learned

1. **The repository interface is the most important boundary in the project.** If placed correctly, every other layer can be tested in isolation. If placed incorrectly (e.g., in the postgres sub-package), the service cannot import it.

2. **The `internal/` boundary is a compile-time feature, not just a convention.** Go refuses to compile if an external package imports anything inside `internal/`. Use this to your advantage — it enforces encapsulation better than any lint rule.

3. **Use `%w` not `%v` when wrapping errors.** `fmt.Errorf("repo: %v", err)` loses the error chain. `fmt.Errorf("repo: %w", err)` preserves it. This is the single most impactful error handling practice in Go.

4. **The partial unique index on `short_code WHERE deleted_at IS NULL` is the key to soft-delete + shortcode reuse.** Without it, either you can't reuse deleted shortcodes (breaking BR-05) or you can't prevent duplicate active shortcodes (breaking BR-01).

5. **The redirect endpoint should use 302 (Found), not 301 (Moved Permanently).** Browsers cache 301s aggressively and never re-request them. This means redirect count won't update on subsequent requests and destination changes won't take effect.

### Common Beginner Mistakes

1. **Putting business logic in handlers** — Handlers should parse HTTP and call the service. They should not validate URLs, generate shortcodes, or execute SQL.
2. **Importing database drivers in the service layer** — The service should never know about `database/sql`, `pgx`, or `pq`. If you import these in service files, refactor immediately.
3. **Circular imports** — If package A imports package B and package B imports package A, Go refuses to compile. Draw the dependency graph before writing code.
4. **Forgetting context propagation** — Every function from handler to repository must accept `ctx context.Context` or database calls won't respect timeouts.
5. **Ignoring soft deletes in read queries** — If queries don't filter `deleted_at IS NULL`, soft-deleted URLs will appear in search results and redirects.
6. **Returning raw database errors to clients** — This leaks internal implementation details (table names, SQL state codes) and can be a security risk. Always map database errors to application-level error types at the repository boundary.
7. **Not writing tests for error paths** — Testing only happy paths gives false confidence. Every success path should have corresponding failure path tests.

### What a Senior Engineer Would Do Differently

1. **Write the repository interface before the implementation.** Define the contract first. The implementation must satisfy the contract by compilation.
2. **Start with the domain model.** The domain model has zero external dependencies and everything else depends on it. Getting it right first prevents expensive refactoring later.
3. **Add the partial unique index from day one.** It's a one-liner in the migration that prevents an entire class of bugs (duplicate shortcodes from race conditions during creation).
4. **Use `UPDATE ... RETURNING` instead of SELECT-then-UPDATE for redirects.** This reduces database round-trips from two to one and eliminates race conditions.
5. **Write integration tests with real PostgreSQL (Docker) from the start.** Don't rely on in-memory mocks for data access testing — they verify the mock, not the database.
6. **Add `golangci-lint` early.** It catches issues (unused imports, variable shadowing) that `go vet` misses and saves time during code review.

### What Should Never Be Over-Engineered

1. **Don't add auth when it's not in scope.** The project doesn't need JWT tokens, OAuth flows, or user accounts until those features are explicitly required.
2. **Don't add caching until it's needed.** A cache introduces invalidation complexity. If the redirect endpoint has sub-50ms latency without a cache, you don't need one.
3. **Don't use an ORM.** Hand-written SQL is explicit, testable, and transparent. GORM's implicit behavior is a debugging nightmare for educational projects.
4. **Don't split into microservices prematurely.** A single Go binary with PostgreSQL is the right architecture until clear bounded-context boundaries emerge.
5. **Don't generate mocks with `mockery` or `go generate`.** Hand-written mocks are easier to understand, maintain, and debug in Go codebases.
6. **Don't add a DI framework.** Manual dependency injection (constructor parameters) is clear, debuggable, and fast. Di frameworks (Uber's dig, Facebook's dig) add complexity without proportional benefit for small projects.

### Final Checklist (Engineering Principles)

- [ ] Keep it simple — prefer the simplest solution that works
- [ ] DRY — don't duplicate code or validation logic
- [ ] KISS — keep interfaces small and focused
- [ ] YAGNI — don't build features you don't need yet
- [ ] Composition over inheritance — Go favors struct embedding and interface composition
- [ ] Dependency inversion — high-level modules depend on abstractions, not concrete types
- [ ] Explicit over implicit — make dependencies and behavior visible
- [ ] Test the behavior, not the implementation
- [ ] Make it work, make it right, make it fast — in that order

---

## Final Project Checklist

### Phase 1 — Bootstrap
- [ ] `go mod init github.com/mmk31585/url-shortener`
- [ ] Directory structure created (`cmd/server/`, all `internal/*/` packages)
- [ ] Placeholder `.go` files in each package
- [ ] `cmd/server/main.go` with minimal `func main()`
- [ ] `Makefile` with build, test, fmt, vet targets
- [ ] `.gitignore` and `.env.example` created
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` succeeds
- [ ] `go test ./...` succeeds (zero tests OK)
- [ ] `gofmt -s .` produces no diffs
- [ ] Committed: `chore: bootstrap Go module and project skeleton`

### Phase 2 — Config & Logger
- [ ] `internal/config/config.go` — `Config` struct with `Load()` function
- [ ] `internal/logger/logger.go` — `New(env string) *slog.Logger`
- [ ] Config reads env vars via `os.LookupEnv` with sensible defaults
- [ ] Logger uses text handler for development, JSON handler for production
- [ ] `main.go` wires config + logger
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` succeeds
- [ ] Committed: `feat: add configuration loading and structured logging`

### Phase 3 — Domain Layer
- [ ] `internal/domain/url.go` — URL struct with all 7 fields and JSON tags
- [ ] `internal/domain/errors.go` — Sentinel errors (ErrURLNotFound, ErrURLAlreadyDeleted, etc.)
- [ ] All domain types have Go doc comments
- [ ] No external dependencies in domain package
- [ ] `go build ./internal/domain/...` succeeds
- [ ] `go vet ./internal/domain/...` is clean
- [ ] Domain tests pass (IsValidStatus, sentinel errors, JSON tags, default values)
- [ ] Coverage > 95%
- [ ] Committed: `feat(domain): add URL entity, ShortCode, status, and domain errors`

### Phase 4 — Storage (PostgreSQL + Goose)
- [ ] Goose dependency added to `go.mod`
- [ ] `migrations/000001_create_urls_table.up.sql` — table with all 7 columns, constraints
- [ ] `migrations/000001_create_urls_table.down.sql` — DROP TABLE
- [ ] `migrations/000002_add_indexes.up.sql` — indexes including partial unique index
- [ ] `migrations/000002_add_indexes.down.sql` — DROP INDEX
- [ ] `internal/storage/storage.go` — `Storage` struct, `New(url string)`, `DB()`, `Close()`
- [ ] `New()` auto-runs Goose migrations on connection
- [ ] DB connection pool configured (MaxOpenConns, MaxIdleConns, ConnMaxLifetime)
- [ ] All integration tests pass against real PostgreSQL (Docker)
- [ ] Committed: `feat(storage): add PostgreSQL connection with Goose auto-migrations`

### Phase 5 — Repository
- [ ] `internal/repository/repository.go` — `URLRepository` interface with 7 methods
- [ ] `internal/repository/postgres/url.go` — `PostgresURLRepository` implementing all methods
- [ ] All queries parameterized (no string interpolation)
- [ ] All queries use `QueryRowContext` / `QueryContext` (context-aware)
- [ ] Database errors mapped to domain errors at repository boundary
- [ ] `*PostgresURLRepository` satisfies `URLRepository` (compiler check)
- [ ] Mock repository works for service testing
- [ ] All 7 repository methods tested with integration tests
- [ ] Soft delete correctly excludes deleted URLs from read operations
- [ ] Coverage > 90%
- [ ] Committed: `feat(repository): add URLRepository interface and PostgresURLRepository`

### Phase 6 — Shortcode Generator
- [ ] `internal/shortener/shortener.go` — `Shortener` interface + `RandomShortener`
- [ ] `Generate()` uses `crypto/rand.Read()` + base62 encoding
- [ ] Generated shortcodes are exactly 8 characters, valid base62
- [ ] Unique across 10K+ iterations (no collisions)
- [ ] `RandomShortener` satisfies `Shortener` interface
- [ ] Coverage 100%
- [ ] Committed: `feat(shortener): add Shortener interface and RandomShortener`

### Phase 7 — Validator
- [ ] `internal/validator/validator.go` — `ValidateURL`, `ValidateShortCode`, `ValidateID`
- [ ] `ValidateURL` checks scheme (http/https), host, length (max 2048)
- [ ] `ValidateShortCode` checks length (exactly 8) and base62 characters
- [ ] `ValidateID` checks id > 0
- [ ] All functions are pure (no I/O, no state, no side effects)
- [ ] All boundary conditions tested (0, 1, 7, 8, 9, 2047, 2048, 2049)
- [ ] Coverage 100%
- [ ] Committed: `feat(validator): add URL, shortcode, and ID validation functions`

### Phase 8 — Service Layer
- [ ] `internal/service/service.go` — `URLService` interface with 7 methods
- [ ] `internal/service/url_service.go` — `urlService` implementing all methods
- [ ] `CreateURL`: validate → generate → persist → return URL
- [ ] `GetURL`: fetch by shortcode, return or error
- [ ] `ListURLs`: return all active URLs (empty slice if none)
- [ ] `UpdateURL`: validate → find → modify → persist → return URL
- [ ] `DeleteURL`: find → soft delete
- [ ] `Redirect`: find → increment count → return original URL
- [ ] `GetStats`: fetch URL with redirect count
- [ ] All 10 business rules (BR-01 to BR-10) enforced
- [ ] Service has zero HTTP or database imports
- [ ] All mocks for service testing (hand-written mock repository)
- [ ] Coverage > 90%
- [ ] Committed: `feat(service): add URLService with all business logic`

### Phase 9 — Handler Layer (HTTP Handlers)
- [ ] `internal/handler/handler.go` — BaseHandler struct
- [ ] `internal/handler/url.go` — All 7 endpoint handlers:
  - [ ] `CreateURL` → POST `/api/v1/urls` → 201 on success
  - [ ] `GetURL` → GET `/api/v1/urls/{shortcode}` → 200 or 404
  - [ ] `Redirect` → GET `/r/{shortcode}` → 302 with Location header
  - [ ] `UpdateURL` → PUT `/api/v1/urls/{shortcode}` → 200 on success
  - [ ] `DeleteURL` → DELETE `/api/v1/urls/{shortcode}` → 204 on success
  - [ ] `GetStats` → GET `/api/v1/urls/{shortcode}/stats` → 200 or 404
  - [ ] `Health` → GET `/health` → 200 with DB status
- [ ] Handlers contain zero business logic, zero SQL, zero net/url calls
- [ ] Consistent error response format `{error: {code, message, request_id}}`
- [ ] request_id from middleware context, included in all error responses
- [ ] 302 redirect (not 301) for all redirects
- [ ] Handler tests using `httptest` for all endpoints (success + all error cases)
- [ ] Coverage > 85%
- [ ] Committed: `feat(handler): add all HTTP handlers with error handling`

### Phase 10 — Router + Middleware
- [ ] `internal/router/router.go` — `NewRouter(svc, logger) *http.ServeMux` with all 8 routes
- [ ] `internal/middleware/middleware.go` — Three middleware functions:
  - [ ] Request ID middleware: UUID v4, context, response header
  - [ ] Logging middleware: method, path, status, duration, request_id
  - [ ] Recovery middleware: catch panics, return 500 generic message
- [ ] Middleware applied in correct order (outer-to-inner)
- [ ] All routes bound correctly in router
- [ ] Middleware tests for all three functions
- [ ] Committed: `feat(middleware): add logging, request ID, and recovery middleware + router`

### Phase 11 — Error Handling Polish
- [ ] `internal/error/error.go` — `AppError` struct with Code, Message, StatusCode
- [ ] Constructor functions: `NewValidationError`, `NewNotFoundError`, `NewConflictError`, `NewGoneError`, `NewInternalError`
- [ ] All errors use `%w` wrapping at every layer boundary
- [ ] `errors.Is` works across all layers for each sentinel error
- [ ] `errors.As` used in handler to extract `*AppError` for HTTP responses
- [ ] Top-level `defer/recover` in `main.go` catches unhandled panics
- [ ] No raw database errors reach HTTP clients
- [ ] Stack traces only in development mode
- [ ] Committed: `refactor(errors): add AppError type and consistent error propagation`

### Phase 12 — Swagger / OpenAPI Documentation
- [ ] `docs/swagger.yaml` — OpenAPI 3.0 specification
- [ ] All 8 endpoints documented with request/response schemas
- [ ] Example request and response bodies for every endpoint
- [ ] Error responses documented for each endpoint
- [ ] Shared components for `URL`, `ErrorResponse`, `StatsResponse`
- [ ] YAML is valid and parseable
- [ ] Committed: `docs: add Swagger/OpenAPI API documentation`

### Phase 13 — Docker & Deployment
- [ ] Multi-stage `Dockerfile` (builder + runtime stages)
- [ ] `CGO_ENABLED=0` for static binary
- [ ] `ldflags="-s -w"` for stripped binary
- [ ] `docker-compose.yml` with App + PostgreSQL
- [ ] `.dockerignore` excluding build artifacts, tests, docs
- [ ] `HEALTHCHECK` in Dockerfile
- [ ] All environment variables set via config (not hardcoded)
- [ ] Full stack works: `docker-compose up` → health check succeeds → create URL works
- [ ] Committed: `feat(docker): add multi-stage Dockerfile and docker-compose`

### Phase 14 — Code Quality & Refactoring
- [ ] `gofmt -s .` produces no diffs
- [ ] `goimports` produces no diffs
- [ ] `go vet ./...` is clean
- [ ] `go mod tidy` has been run
- [ ] No duplicate code (DRY)
- [ ] No unnecessary abstractions (KISS)
- [ ] No speculative features (YAGNI)
- [ ] No `init()` functions (except `sync.Once`)
- [ ] No global variables (except sentinel errors)
- [ ] No panics (except recovered in main)
- [ ] No unused imports or variables
- [ ] Go doc comments on all exported symbols
- [ ] Package-level doc comments on every package
- [ ] Import statements grouped correctly
- [ ] No circular imports
- [ ] `make all` passes (fmt → vet → build → test)
- [ ] Committed: `refactor: code quality review and cleanup`

### Phase 15 — Final Testing & Documentation
- [ ] `go test -race -cover ./...` passes with > 80% coverage
- [ ] Every exported function has at least one test
- [ ] All error paths tested for every handler and service method
- [ ] Business rules BR-01 to BR-10 validated by tests
- [ ] README.md accurately reflects the implemented project
- [ ] All doc comments are complete
- [ ] No TODO/FIXME comments without tracking issues
- [ ] Committed: `test: final test suite and `docs: finalize README for accuracy`

### Phase 16 — Release v1.0.0
- [ ] All previous phases complete and committed
- [ ] `go build -o url-shortener ./cmd/server` produces a working binary
- [ ] Final manual smoke test of all endpoints (create, redirect, update, delete, stats, health)
- [ ] `make all` passes
- [ ] Docker build and compose work correctly
- [ ] Create and push tag: `git tag -a v1.0.0 -m "v1.0.0: Initial stable release"`
- [ ] Committed: `chore: bump version to v1.0.0`
