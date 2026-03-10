# Docker Compose Development Environment Design

**Goal:** Add a local development environment for the Go interview crawler service where Docker Compose manages MySQL and Adminer, while the Go service runs on the host and can read configuration from a root `.env` file.

## Scope

- Add a root `docker-compose.dev.yml` for MySQL and Adminer.
- Add a root `.env.example` with both Compose variables and host-run Go service variables.
- Add lightweight `.env` loading support for the Go service.
- Update developer documentation for the local workflow.

## Non-Goals

- Do not containerize the Go service in this task.
- Do not add production deployment files.
- Do not add init SQL mounts because the service already auto-applies migrations.
- Do not add Redis, mail tooling, or observability containers.

## Recommended Approach

Use a minimal infrastructure-only Compose file plus a root `.env.example`, and teach the Go service to load the root `.env` before reading environment variables. This keeps the dev workflow short:

1. Copy `.env.example` to `.env`
2. Start `mysql` and `adminer` with Compose
3. Run `cd server && go run ./cmd/server`

This avoids duplicating runtime config across shell exports, IDE run configs, and Compose definitions.

## Services

### `mysql`

- Image: `mysql:8.4`
- Exposed port: `3306`
- Named volume for persistence
- `mysqladmin ping` health check
- Environment variables sourced from `.env`

### `adminer`

- Image: `adminer:4`
- Exposed port: `8081`
- Depends on `mysql`
- No persistence required

## Configuration Model

### Root `.env.example`

Should include:

- Compose-side variables:
  - `MYSQL_ROOT_PASSWORD`
  - `MYSQL_DATABASE`
  - `MYSQL_USER`
  - `MYSQL_PASSWORD`
  - `MYSQL_PORT`
  - `ADMINER_PORT`
- Host-run Go service variables:
  - `INTERVIEW_CRAWLER_HTTP_ADDR`
  - `INTERVIEW_CRAWLER_MYSQL_DSN`
  - `INTERVIEW_CRAWLER_LLM_API_KEY`
  - `INTERVIEW_CRAWLER_LLM_BASE_URL`
  - `INTERVIEW_CRAWLER_LLM_MODEL`
  - scheduler variables as examples

### Go `.env` Loading

- Load the root `.env` automatically before reading config fields.
- Do not overwrite environment variables already set by the shell or IDE.
- If no `.env` file exists, continue normally with existing behavior.

## Files

- Create `docker-compose.dev.yml`
- Create `.env.example`
- Modify `server/internal/config/config.go`
- Add tests for `.env` loading behavior
- Update `server/README.md`

## Verification Strategy

- `docker compose -f docker-compose.dev.yml config`
- `docker compose -f docker-compose.dev.yml up -d`
- `docker compose -f docker-compose.dev.yml ps`
- `docker compose -f docker-compose.dev.yml exec mysql mysqladmin ping -h localhost -uroot -p"$MYSQL_ROOT_PASSWORD"`
- `cd server && go test ./...`
- `cd server && go build -buildvcs=false ./...`

## Risks

- If the host already uses port `3306` or `8081`, the developer must override the exposed port in `.env`.
- `.env` parsing should stay minimal and only target local developer convenience.
- Compose health checks do not guarantee host-side readiness beyond basic MySQL availability, but they are sufficient for local bootstrapping.
