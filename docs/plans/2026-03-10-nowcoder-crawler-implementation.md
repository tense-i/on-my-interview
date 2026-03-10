# Go Interview Crawler Service Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go HTTP service under `server/` that crawls NowCoder interview posts, deduplicates raw posts in MySQL, extracts structured interview data via a pluggable LLM interface, and exposes job/post/question query APIs.

**Architecture:** The first version is an asynchronous monolith. HTTP handlers and the scheduler only create jobs. A background worker runs crawl jobs, persists deduplicated raw posts, invokes LLM extraction on new or changed posts, and materializes searchable question rows in MySQL.

**Tech Stack:** Go 1.26, Chi router, standard library HTTP client, `go-sql-driver/mysql`, MySQL 8, cron scheduling via `robfig/cron/v3`, OpenAI-compatible JSON HTTP API.

---

### Task 1: Bootstrap the Go service module

**Files:**
- Create: `server/go.mod`
- Create: `server/cmd/server/main.go`
- Create: `server/internal/config/config.go`
- Create: `server/internal/app/app.go`
- Create: `server/internal/http/router.go`
- Create: `server/internal/http/handlers/health.go`
- Create: `server/internal/http/handlers/jobs.go`
- Create: `server/internal/http/handlers/posts.go`
- Create: `server/internal/http/handlers/questions.go`
- Create: `server/internal/http/handlers/companies.go`
- Create: `server/internal/storage/mysql/mysql.go`
- Create: `server/internal/testutil/testenv.go`
- Test: `server/internal/http/router_test.go`

**Step 1: Write the failing test**

Add a router test that expects `GET /health` to return status `200` and a JSON body containing `"status":"ok"`.

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/http -run TestRouterHealth -v`
Expected: FAIL because the module or router code does not exist yet.

**Step 3: Write minimal implementation**

- Initialize the module and dependencies.
- Add config loading from environment.
- Wire an app struct with DB, extractor, crawler registry, job service, and HTTP router placeholders.
- Implement the `/health` handler and base router.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/http -run TestRouterHealth -v`
Expected: PASS

### Task 2: Add schema and repository support for jobs, raw posts, parse results, and questions

**Files:**
- Create: `server/migrations/001_init.sql`
- Create: `server/internal/storage/repository/models.go`
- Create: `server/internal/storage/repository/repository.go`
- Create: `server/internal/storage/mysql/repository.go`
- Test: `server/internal/storage/mysql/repository_test.go`

**Step 1: Write the failing test**

Add repository tests that verify:
- inserting a raw post creates one row keyed by `(platform, source_post_id)`
- inserting the same raw post again is detected as unchanged
- updating the edited time or content hash marks the post as updated

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/storage/mysql -run TestUpsertRawPost -v`
Expected: FAIL because repository code does not exist.

**Step 3: Write minimal implementation**

- Define SQL schema.
- Implement MySQL repository interfaces for jobs, raw posts, parse results, questions, and job-post mapping.
- Implement upsert logic and disposition reporting.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/storage/mysql -run TestUpsertRawPost -v`
Expected: PASS

### Task 3: Implement the NowCoder crawler and normalization layer

**Files:**
- Create: `server/internal/crawler/types.go`
- Create: `server/internal/crawler/registry.go`
- Create: `server/internal/crawler/nowcoder/client.go`
- Create: `server/internal/crawler/nowcoder/client_test.go`

**Step 1: Write the failing test**

Add a test with a mocked HTTP server that returns the current NowCoder search response shape and verify normalized posts contain platform, source ID, title, content, source timestamps, and raw payload.

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/crawler/nowcoder -run TestSearchPosts -v`
Expected: FAIL because the crawler does not exist.

**Step 3: Write minimal implementation**

- Add the crawler interface and registry.
- Implement `NowCoderCrawler.SearchPosts`.
- Normalize search records into a common `RawPostInput`.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/crawler/nowcoder -run TestSearchPosts -v`
Expected: PASS

### Task 4: Implement the LLM extractor contract and OpenAI-compatible provider

**Files:**
- Create: `server/internal/llm/types.go`
- Create: `server/internal/llm/schema.go`
- Create: `server/internal/llm/openai/extractor.go`
- Create: `server/internal/llm/openai/extractor_test.go`

**Step 1: Write the failing test**

Add a test with a mocked OpenAI-compatible HTTP endpoint that returns JSON content. Verify the extractor validates the schema and returns a structured post with flattened question rows.

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/llm/openai -run TestExtractorExtract -v`
Expected: FAIL because extractor code does not exist.

**Step 3: Write minimal implementation**

- Define shared structured schema types.
- Create the prompt and HTTP client for an OpenAI-compatible API.
- Parse and validate JSON output.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/llm/openai -run TestExtractorExtract -v`
Expected: PASS

### Task 5: Implement job service, worker, scheduler, and reparse flow

**Files:**
- Create: `server/internal/jobs/service.go`
- Create: `server/internal/jobs/worker.go`
- Create: `server/internal/jobs/scheduler.go`
- Create: `server/internal/jobs/service_test.go`

**Step 1: Write the failing test**

Add a service test that creates a job, runs the worker with mocked crawler and extractor dependencies, and verifies:
- new posts are inserted
- unchanged posts are skipped
- changed posts are reparsed
- failed extraction marks parse status as failed

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/jobs -run TestWorkerRunJob -v`
Expected: FAIL because services do not exist.

**Step 3: Write minimal implementation**

- Implement job creation and listing.
- Implement background worker execution.
- Add optional cron-based scheduling.
- Add single-post reparse support.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/jobs -run TestWorkerRunJob -v`
Expected: PASS

### Task 6: Implement query APIs for jobs, posts, questions, and companies

**Files:**
- Modify: `server/internal/http/router.go`
- Modify: `server/internal/http/handlers/jobs.go`
- Modify: `server/internal/http/handlers/posts.go`
- Modify: `server/internal/http/handlers/questions.go`
- Modify: `server/internal/http/handlers/companies.go`
- Create: `server/internal/http/handlers/common.go`
- Create: `server/internal/http/handlers/api_test.go`

**Step 1: Write the failing test**

Add API tests that verify:
- creating a crawl job returns `202`
- listing posts filters by platform/company/tag
- fetching a single post returns raw and parsed content
- listing questions returns per-question rows
- listing companies returns aggregated counts

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/http/handlers -run TestAPI -v`
Expected: FAIL because handlers are incomplete.

**Step 3: Write minimal implementation**

- Implement request/response types.
- Add filtering and pagination query support.
- Add reparse endpoint wiring.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/http/handlers -run TestAPI -v`
Expected: PASS

### Task 7: Verify the full server and document operations

**Files:**
- Create: `server/README.md`
- Modify: `memory/FACTS.md`
- Modify: `memory/EXPERIENCE.md` if a reusable implementation pitfall emerges

**Step 1: Write the failing test**

Add or extend tests to cover a full in-memory job execution path and verify the documented sample configuration boots the app cleanly.

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./...`
Expected: FAIL until missing wiring is complete.

**Step 3: Write minimal implementation**

- Add operational README and environment variable docs.
- Update stable memory only if the new server layout and commands are real.

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./...`
Expected: PASS

### Task 8: Final verification

**Files:**
- Verify: `server/...`
- Verify: `docs/plans/2026-03-10-nowcoder-crawler-context-manifest.md`
- Verify: `docs/plans/2026-03-10-nowcoder-crawler-design.md`
- Verify: `docs/plans/2026-03-10-nowcoder-crawler-implementation.md`

**Step 1: Run the complete verification suite**

Run:
- `cd server && go test ./...`
- `cd server && go build ./...`

**Step 2: Confirm final requirements**

Verify the implementation includes:
- multi-platform-ready raw post schema
- MySQL dedupe and repeated-crawl handling
- pluggable LLM extractor with OpenAI-compatible default
- scheduled and manual crawl jobs
- query APIs for jobs, posts, questions, and companies
