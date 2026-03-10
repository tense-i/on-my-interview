# LLM Usage Window Aggregation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a restart-safe, MySQL-backed 12-hour Asia/Shanghai aggregate for LLM token usage and estimated cost, and expose finalized usage windows through a read API.

**Architecture:** The extractor will decode provider `usage`, the job service will forward observed usage to the repository, and the MySQL repository will aggregate into one `llm_usage_windows` row per Beijing half-day. Expired aggregating rows will be finalized during both write and read paths so the API returns settled facts without a separate scheduler.

**Tech Stack:** Go 1.24.6, MySQL 8.4, Chi HTTP router, DeepSeek official pricing, standard library time/math/JSON handling.

---

### Task 1: Extend LLM types to carry provider usage

**Files:**
- Modify: `server/internal/llm/types.go`
- Modify: `server/internal/llm/openai/extractor.go`
- Modify: `server/internal/llm/openai/extractor_test.go`

**Step 1: Write the failing test**

Extend extractor tests to assert that a completion response with a `usage` object is decoded into the returned structured result.

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/llm/openai -run TestExtractorExtract -count=1`
Expected: FAIL because usage is not exposed yet.

**Step 3: Write minimal implementation**

- Add usage structs to `server/internal/llm/types.go`
- Add a `Usage` field to the extractor result model
- Decode `usage` in `server/internal/llm/openai/extractor.go`

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/llm/openai -run TestExtractorExtract -count=1`
Expected: PASS

### Task 2: Add repository models and window math/cost helpers

**Files:**
- Modify: `server/internal/storage/repository/models.go`
- Create: `server/internal/usage/window.go`
- Create: `server/internal/usage/window_test.go`

**Step 1: Write the failing test**

Add tests for:
- selecting the correct Beijing half-day window
- computing DeepSeek `deepseek-chat` estimated cost from usage counters

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/usage -count=1`
Expected: FAIL because the package/helpers do not exist yet.

**Step 3: Write minimal implementation**

- Define usage window and totals models in the repository package
- Implement `server/internal/usage/window.go` with:
  - Asia/Shanghai window selection
  - DeepSeek cost calculation
  - `DECIMAL(20,8)` friendly rounding

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/usage -count=1`
Expected: PASS

### Task 3: Add MySQL schema and aggregate repository methods

**Files:**
- Modify: `server/internal/storage/mysql/migrations/001_init.sql`
- Modify: `server/internal/storage/repository/repository.go`
- Modify: `server/internal/storage/mysql/repository.go`
- Modify: `server/internal/storage/mysql/repository_test.go`

**Step 1: Write the failing test**

Add repository tests that verify:
- observed usage creates/updates one aggregate row for the current window
- a second write in the same window increments totals
- expired `aggregating` rows become `finalized`
- finalized windows can be listed with total summaries

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/storage/mysql -run 'TestRecordUsageWindow|TestListUsageWindows' -count=1`
Expected: FAIL because the table and repository methods do not exist yet.

**Step 3: Write minimal implementation**

- Add `llm_usage_windows` migration
- Add repository methods for:
  - recording observed usage into the current window
  - finalizing expired windows
  - listing finalized windows and range totals
- Keep updates transactionally safe

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/storage/mysql -run 'TestRecordUsageWindow|TestListUsageWindows' -count=1`
Expected: PASS

### Task 4: Hook usage aggregation into the job flow

**Files:**
- Modify: `server/internal/jobs/service.go`
- Modify: `server/internal/jobs/service_test.go`

**Step 1: Write the failing test**

Add a job service test proving that a successful parse with observed usage records a usage window update.

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/jobs -run TestRunJobRecordsUsage -count=1`
Expected: FAIL because the service does not forward usage yet.

**Step 3: Write minimal implementation**

- After a successful extraction, if usage is present, call the repository aggregation method with:
  - provider name
  - model name
  - observed usage
  - current time

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/jobs -run TestRunJobRecordsUsage -count=1`
Expected: PASS

### Task 5: Expose the finalized usage API

**Files:**
- Modify: `server/internal/http/handlers/common.go`
- Create: `server/internal/http/handlers/usage.go`
- Modify: `server/internal/http/handlers/api_test.go`
- Modify: `server/internal/http/router.go`

**Step 1: Write the failing test**

Add handler tests for:
- `GET /api/v1/usage/windows`
- time range filtering and totals payload

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/http/handlers -run TestListUsageWindows -count=1`
Expected: FAIL because the handler and route do not exist yet.

**Step 3: Write minimal implementation**

- Extend the query interface with usage-window listing
- Add `usage.go` handler and response structs
- Register the route in `server/internal/http/router.go`

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/http/handlers -run TestListUsageWindows -count=1`
Expected: PASS

### Task 6: Full verification and docs refresh

**Files:**
- Modify: `server/README.md`
- Modify: `memory/FACTS.md`
- Modify: `memory/EXPERIENCE.md` if new stable pitfalls appear

**Step 1: Run targeted tests**

Run:
- `cd server && go test ./internal/llm/openai -count=1`
- `cd server && go test ./internal/usage -count=1`
- `cd server && go test ./internal/storage/mysql -count=1`
- `cd server && go test ./internal/jobs -count=1`
- `cd server && go test ./internal/http/handlers -count=1`

**Step 2: Run required Go 1.26 fix pass**

Run:
- `cd server && GOTOOLCHAIN=go1.26.0 go fix -diff ./...`
- `cd server && GOTOOLCHAIN=go1.26.0 go fix ./...`

**Step 3: Run full suite**

Run:
- `cd server && go test ./...`
- `cd server && go build -buildvcs=false ./...`

**Step 4: Update docs**

- Document the new usage API and billing semantics in `server/README.md`
- Record stable project facts in `memory/FACTS.md`
