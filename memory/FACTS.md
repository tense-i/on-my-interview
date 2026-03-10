# Project Facts

## Context Engineering Workflow
- This repository uses a traceable Context Engineering workflow: every task must leave behind a Context Manifest under `docs/plans/`.
- The authoritative repository context layout is:
  - `rules/` for minimal always-on rules
  - `memory/` for long-lived facts/experience/decisions
  - `skills/` for task-triggered playbooks
  - `docs/` for on-demand background material
- `today.md` is the repository scratchpad for temporary goals, hypotheses, commands, and raw findings. Stable conclusions should be distilled into `memory/*`.

## Current Repository Shape
- The repository contains a Python prototype crawler under `newcoder-crawler/` and a Go service under `server/`.
- The Go backend module path is `on-my-interview/server`.
- `server/go.mod` currently targets Go `1.24.6`.
- The first implemented platform crawler is `server/internal/crawler/nowcoder`.
- The default LLM backend is an OpenAI-compatible HTTP client at `server/internal/llm/openai`.
- Local infrastructure for development is defined in root `docker-compose.dev.yml`.
- Root `.env.example` is the template for both Compose variables and host-run Go service variables.

## Backend Runtime Surface
- Main entrypoint: `server/cmd/server/main.go`
- App wiring: `server/internal/app/app.go`
- Environment config loader: `server/internal/config/config.go`
- HTTP router: `server/internal/http/router.go`
- MySQL repository: `server/internal/storage/mysql`
- Embedded schema migration: `server/migrations/001_init.sql`
- The config loader auto-loads the nearest parent `.env` file without overriding already-set process environment variables.

## Storage Invariants
- Raw posts are deduplicated by `(platform, source_post_id)`.
- `raw_posts` keeps the original payload JSON and source timestamps.
- Structured parse results are stored separately in `post_parse_results`.
- Interview questions follow the minimal-line rule: one distinct question per row in `interview_questions`.
- `question_tags` is maintained from parsed question tags for later graph-style analysis.
- `llm_usage_windows` stores global LLM usage and estimated cost aggregated into fixed 12-hour `Asia/Shanghai` windows.
- The usage query surface returns finalized windows only; in-progress aggregation checkpoints are persisted but not exposed.

## API Surface
- `GET /health`
- `POST /api/v1/crawl/jobs`
- `GET /api/v1/crawl/jobs`
- `GET /api/v1/crawl/jobs/:id`
- `GET /api/v1/posts`
- `GET /api/v1/posts/:platform/:sourcePostID`
- `POST /api/v1/posts/:platform/:sourcePostID/reparse`
- `GET /api/v1/questions`
- `GET /api/v1/companies`
- `GET /api/v1/usage/windows`

## Configuration Surface
- Required:
  - `INTERVIEW_CRAWLER_MYSQL_DSN`
  - `INTERVIEW_CRAWLER_LLM_API_KEY`
- Optional:
  - `INTERVIEW_CRAWLER_HTTP_ADDR`
  - `INTERVIEW_CRAWLER_NOWCODER_BASE_URL`
  - `INTERVIEW_CRAWLER_LLM_BASE_URL`
  - `INTERVIEW_CRAWLER_LLM_MODEL`
  - `INTERVIEW_CRAWLER_SCHEDULE_ENABLED`
  - `INTERVIEW_CRAWLER_SCHEDULE_INTERVAL`
  - `INTERVIEW_CRAWLER_SCHEDULE_PLATFORMS`
  - `INTERVIEW_CRAWLER_SCHEDULE_KEYWORDS`
  - `INTERVIEW_CRAWLER_SCHEDULE_PAGES`
- DeepSeek can be used through the existing OpenAI-compatible extractor with:
  - `INTERVIEW_CRAWLER_LLM_BASE_URL=https://api.deepseek.com`
  - `INTERVIEW_CRAWLER_LLM_MODEL=deepseek-chat`
  - The configured base URL must not already include `/v1`, because the extractor appends `/v1/chat/completions`.

## Known Good Commands
- Run backend tests: `cd server && go test ./...`
- Build backend: `cd server && go build -buildvcs=false ./...`
- Run backend: `cd server && go run ./cmd/server`
- Render dev Compose config: `docker compose -f docker-compose.dev.yml --env-file .env.example config`
- Start dev MySQL and Adminer: `docker compose -f docker-compose.dev.yml up -d`
- Stop dev MySQL and Adminer: `docker compose -f docker-compose.dev.yml down`
- Inspect Go fix changes when Go 1.26 toolchain is available: `cd server && GOTOOLCHAIN=go1.26.0 go fix -diff ./...`
- Apply Go fix changes when Go 1.26 toolchain is available: `cd server && GOTOOLCHAIN=go1.26.0 go fix ./...`
