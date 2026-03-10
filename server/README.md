# Go Interview Crawler Service

This service crawls interview-experience posts from NowCoder, deduplicates raw posts in MySQL, extracts structured interview information through a pluggable LLM interface, and exposes HTTP query APIs.

## Features

- Multi-platform-ready raw post schema keyed by `(platform, source_post_id)`
- NowCoder crawler implementation via the current search API
- MySQL persistence for crawl jobs, raw posts, structured parse results, question rows, and tags
- OpenAI-compatible extractor as the default LLM backend
- 12-hour Asia/Shanghai aggregate usage and cost tracking for observed LLM responses
- Built-in background job runner and interval-based scheduler
- Query APIs for jobs, posts, questions, and company aggregates

## Required Environment Variables

- `INTERVIEW_CRAWLER_MYSQL_DSN`
  - Example: `user:password@tcp(127.0.0.1:3306)/crawler?parseTime=true&charset=utf8mb4`
- `INTERVIEW_CRAWLER_LLM_API_KEY`
- `INTERVIEW_CRAWLER_LLM_BASE_URL`
  - Default: `https://api.openai.com`
- `INTERVIEW_CRAWLER_LLM_MODEL`
  - Default: `gpt-4.1-mini`

## Optional Environment Variables

- `INTERVIEW_CRAWLER_HTTP_ADDR`
  - Default: `:8080`
- `INTERVIEW_CRAWLER_NOWCODER_BASE_URL`
  - Default: `https://gw-c.nowcoder.com`
- `INTERVIEW_CRAWLER_SCHEDULE_ENABLED`
  - Default: `false`
- `INTERVIEW_CRAWLER_SCHEDULE_INTERVAL`
  - Default: `1h`
- `INTERVIEW_CRAWLER_SCHEDULE_PLATFORMS`
  - Default: `nowcoder`
- `INTERVIEW_CRAWLER_SCHEDULE_KEYWORDS`
  - Default: `面经`
- `INTERVIEW_CRAWLER_SCHEDULE_PAGES`
  - Default: `1`

## Run

```bash
cd server
go run ./cmd/server
```

The service pings MySQL and auto-applies the embedded schema migration on startup.

## Local Dev with Docker Compose

1. Create a local environment file at the repository root:

```bash
cp .env.example .env
```

2. Start local infrastructure from the repository root:

```bash
docker compose -f docker-compose.dev.yml up -d
```

3. Run the Go service on the host:

```bash
cd server
go run ./cmd/server
```

4. Check local services:

- MySQL: `127.0.0.1:3306`
- Adminer: `http://localhost:8081`
- API health: `http://localhost:8080/health`

Notes:

- The config package automatically loads the repository root `.env` file when variables are not already set in the shell.
- If you change `MYSQL_PORT` in `.env`, update `INTERVIEW_CRAWLER_MYSQL_DSN` to match because the `.env` loader does not expand variable references inside values.
- The Go service still runs on the host in this workflow; Compose manages only MySQL and Adminer.

## Main Endpoints

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

## Usage Billing Windows

- Usage is aggregated globally into fixed Asia/Shanghai windows:
  - `00:00-12:00`
  - `12:00-24:00`
- The API returns finalized windows only. In-progress half-day checkpoints are persisted internally but are not exposed.
- The first cost calculator is implemented for `deepseek-chat`.
- Cost is estimated from provider `usage` counters using the current DeepSeek pricing model configured in code.

Example:

```bash
curl 'http://localhost:8080/api/v1/usage/windows?limit=10'
```

## Example Crawl Request

```json
{
  "platforms": ["nowcoder"],
  "keywords": ["面经", "后端 面经"],
  "pages": 2,
  "force_reparse": false
}
```

## Current Notes

- The first crawler implementation uses the NowCoder search API and does not fetch HTML detail pages.
- Query APIs read from MySQL directly; there is no Redis or message queue in this version.
- The scheduler uses a fixed duration interval, not cron syntax, in this first version.
