# Modular Monolith Clean Architecture Boilerplate

Base project for building REST APIs in Go using a modular-monolith layout that enforces Clean Architecture boundaries, explicit dependency direction, and automated module discovery.

## Key Features

- **Clean Architecture layers** (entity → repository → usecase → handler) per module.
- **Auto-wired modules** via registry + DI container; new modules register themselves.
- **Environment-driven config** including DB + security settings.
- **PostgreSQL infrastructure** with migrations, seeds, and connection helper.
- **Utilities**: JWT & inter-service middleware, password hashing helper, health checks.
- **Scaffolding tooling**: module generator, migrations runner, seeder, rename helper.

## Project Structure

```
cmd/
  server/          # API entrypoint (bootstraps router, DI container, modules)
  migration/       # go run ./cmd/migration <up|down>
  seed/            # go run ./cmd/seed (applies db/seeds)
  rename/          # search/replace module paths safely
  generate/        # module scaffolder
db/
  migrations/      # SQL files with -- +migrate directives
  seeds/           # SQL seed scripts
internal/
  config/          # env loading + structs
  infrastructure/  # cross-cutting adapters (e.g., database)
  modules/
    auth/          # Auth boundary (login/refresh, token mgmt)
    registry/      # module registry + filesystem scanner
    user/          # sample module (entity/usecase/repository/handler)
  server/
    container/     # DI container for shared deps
    healthcheck/   # /healthz, /readyz handlers
    middleware/    # JWT + inter-service middlewares
    httpresp/      # response helpers
    security/      # password hashing utilities
  storage/
    storage.go     # pluggable local/S3 backend
    uploader/      # base64 upload helper with validation
```

## Prerequisites

- Go 1.25+
- PostgreSQL instance
- Docker & Docker Compose (optional, for containerized development)

## Configuration

Copy `.env.example` to `.env` (or export variables another way) and set:

```
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=modmonolith
DB_SSL_MODE=disable
JWT_SECRET=changeme
INTER_SERVICE_TOKEN=internal-token
BCRYPT_COST=10
```

The application reads these via `internal/config`.

## Docker Development

For containerized development with PostgreSQL and optional MinIO:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

Services:
- **App**: API server on port 8080
- **DB**: PostgreSQL on port 5432
- **MinIO**: S3-compatible storage on ports 9000 (API) and 9001 (console)

## Make Targets

| Command | Description |
| --- | --- |
| `make run` | Start API (`go run ./cmd/server`). |
| `make build` | Build binary into `bin/server`. |
| `make test` | Run `go test ./...`. |
| `make migrate-up` | Apply pending migrations. |
| `make migrate-down` | Roll back applied migrations (all, newest first). |
| `make migrate-seed` | Run seed scripts from `db/seeds`. |
| `make rename old=<old> new=<new>` | Replace module paths across the repo. |
| `make generate modules <name>` | Scaffold a new module (see below). |

## Database Migrations & Seeds

- **Apply:** `go run ./cmd/migration up`
- **Rollback:** `go run ./cmd/migration down`
- **Seed:** `go run ./cmd/seed`

Migrations must contain `-- +migrate Up/Down` markers; seeding splits SQL on `;` and executes sequentially.

## Module Generator

Create a new module (e.g., `order`):

```bash
make generate modules order
```

This will:

1. Create `internal/modules/order` with entity, repository (interface + Postgres impl), usecase, handler, tests, and module registration file.
2. Auto-import the module inside `internal/modules/modules.go` so it is discovered at runtime.
3. Auto-format all generated files.

Modules self-register with the registry in their `module.go`, so the HTTP router loads them automatically.

## Runtime Flow

1. `cmd/server` loads config, opens DB, builds the DI container.
2. Health routes + global middleware (logging, JWT, inter-service token) are set up.
3. `internal/modules.RegisterAll` scans `internal/modules/*`, constructs each module via its builder, and mounts their routes under `/api/v1`.
4. Requests hit handlers → usecase → repository according to dependency direction.

## Testing

Run the full suite:

```bash
make test
```

Unit tests live close to their usecases/handlers; focus on business logic (usecase layer) with repository stubs to keep tests deterministic.

## Extending the Boilerplate

- **Add more middleware** under `internal/server/middleware` and register in `cmd/server`.
- **Add more infrastructure adapters** (e.g., cache) in `internal/infrastructure` and surface them via the DI container.
- **New modules**: use the generator, add migrations/seeds as needed, expose routes through the existing interface.
- **Static uploads**: use the `storage` package (local uploads or S3/MinIO) and fetch base64 helpers from `internal/storage/uploader`. Serve local files via `/cdn` endpoint on the API server.
- **Auth boundary**: the `auth` module exposes `/auth/login` and `/auth/refresh` endpoints, issuing/refreshing JWT pairs based on the shared user repository and configurable secrets.

## Security & Networking Notes

- Requests pass through XSS sanitization, rate limiting, JWT validation (with configurable TTL), inter-service token checks, and strict security headers.
- CORS is centrally configured via environment variables; tweak allowed origins/methods/headers without code changes.
- If no upstream WAF is present, these defaults provide a reasonable baseline; consider hardening further per deployment needs.

This setup keeps dependencies flowing inward and lets you peel modules into microservices later by extracting their directory as-is.
