# Bitik Backend

Phase 0 backend foundation for the Bitik marketplace.

## Stack

- Go + Gin
- PostgreSQL, Redis, RabbitMQ
- MinIO for local object storage
- OpenSearch for local search
- Jaeger-compatible OTLP tracing target
- Mailpit and MockServer for local mail/SMS stubs

## Local Development

Start the local dependency stack:

```sh
make compose-up
```

Local Postgres is bound to `127.0.0.1:55432` to avoid clashing with a system Postgres on `5432`.

### Docker Compose (without Make)

From `backend/` (same directory as `docker-compose.yml`):

```sh
docker compose up -d
docker compose ps
docker compose logs -f postgres
docker compose down
```

If Compose warns about containers from an old project name, run once:

```sh
docker compose up -d --remove-orphans
```

### Running the API

All `make` targets below are meant to be run from **`backend/`** (the directory that contains this `README.md` and `go.mod`).

| Command | What it does |
| --------|-------------- |
| `make run` | Starts the API using the current **`.env`** (and optional `internal/.env` overrides). Does not copy any file. By default runs **Goose migrations up** once Postgres is reachable (`BITIK_DATABASE_AUTO_MIGRATE`, default `true`). |
| `make run-local` | Requires **`.env.local`**. Copies `.env.local` → **`.env`** (overwrites), then starts the API. Use for day-to-day local work. |
| `make run-production` | Requires **`.env.prod`**. Copies `.env.prod` → **`.env`** (overwrites), then starts the API. Use only when you intend to run with production-like variables. |

You can also start the API without Make, from the monorepo root, if **`go.work`** is present at the repo root:

```sh
go run ./backend/cmd/api
```

In that case the working directory is the monorepo root; dotenv still resolves **`backend/.env`** and **`backend/internal/.env`** (see `internal/config`).

### Environment files

Typical layout under `backend/`:

| File | Role |
| -----|------|
| **`.env.example`** | Committed template; copy to `.env.local` or `.env.prod` and fill in values. |
| **`.env.local`** | Your local secrets and URLs (gitignored). Use with `make run-local`. |
| **`.env.prod`** | Production-oriented values (gitignored). Use with `make run-production`. |
| **`.env`** | The file the process actually loads after a copy step, or when you edit it by hand (gitignored). |
| **`internal/.env`** | Optional extra overrides loaded after `.env` (gitignored). |

Load order (later files override earlier ones for the same variable):

- From **`backend/`**: `.env`, then `internal/.env`.
- From the **monorepo root**: `backend/.env`, then `backend/internal/.env`.

`BITIK_*` variables set in the shell or your IDE still override dotenv (Viper + env).

Useful endpoints:

- `GET /health`
- `GET /ready`
- `GET /metrics` (on the API server by default; set `BITIK_HTTP_METRICS_ADDR` to serve metrics only on that address)
- `GET /version`
- `GET /openapi.yaml` (embedded spec, not cwd-dependent)
- `GET /docs` (Swagger UI, loads CDN assets)
- `GET /swagger` (redirects to `/docs`)

Tracing: set `BITIK_OBSERVABILITY_TRACING_ENABLED=true` and point `BITIK_OBSERVABILITY_OTLP_ENDPOINT` at the local Jaeger OTLP gRPC port (`4317` in Compose).

## Checks

```sh
make test
make test-integration
make test-coverage
make fmt-check
make lint
make build
docker compose config
```

## Database

```sh
make migrate-up      # apply Goose migrations
make migrate-status  # show migration status
make seed            # apply repeatable development/staging seed data
make db-check        # migrate + status + seed
make sqlc-generate   # regenerate internal/store/* from db/queries/*
make sqlc-check      # ensure generated sqlc code is in sync
make migration-create name=create_example_table
```

## Load Tests

Run Phase 12 load scripts with k6:

```sh
bash ./scripts/run-loadtests.sh
```

API conventions are documented in `docs/API_CONVENTIONS.md`.
Database layout, table reference, and rollback policy are documented in `docs/DATABASE.md`.
