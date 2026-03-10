## Task

Add LLM usage and cost instrumentation that aggregates globally into fixed 12-hour Asia/Shanghai windows, persists checkpointed window totals in MySQL, and exposes a factual query API that only returns finalized windows.

## Loaded Context

- `AGENTS.md`
  - Context manifest requirement, brainstorming-before-implementation rule, verification and Go 1.26 `go fix` requirement.
- `memory/FACTS.md`
  - Current backend layout, storage invariants, DeepSeek OpenAI-compatible configuration.
- `memory/EXPERIENCE.md`
  - Existing DeepSeek schema contract pitfall and NowCoder empty-record filtering lesson.
- `server/internal/llm/openai/extractor.go`
  - Current OpenAI-compatible request/response handling where provider `usage` must be observed.
- `server/internal/jobs/service.go`
  - Main parse flow where successful LLM extractions are produced.
- `server/internal/storage/repository/models.go`
  - Existing repository model surface and naming conventions.
- `server/internal/storage/mysql/repository.go`
  - Current MySQL repository style and query patterns.
- `server/internal/http/handlers/*.go`
  - Existing API handler and response style.
- `server/internal/http/router.go`
  - Route registration layout.
- `server/internal/storage/mysql/migrations/001_init.sql`
  - Current schema baseline.
- DeepSeek official docs
  - `https://api-docs.deepseek.com/zh-cn/quick_start/pricing`
  - `https://api-docs.deepseek.com/zh-cn/api/create-chat-completion/`
  - `https://api-docs.deepseek.com/zh-cn/quick_start/token_usage`

## Excluded Context

- Python prototype under `newcoder-crawler/`
  - Not relevant to the Go backend feature.
- Frontend/UI context
  - This task is backend-only.
- Unrelated docs under `docs/`
  - Not needed for the feature design or implementation.

## Assumptions

- Only global aggregate usage/cost is required, not per-request or per-post billing detail.
- Query consumers want finalized, factual 12-hour windows only; in-progress windows are not exposed.
- The first provider-specific cost calculator only needs to support DeepSeek `deepseek-chat`.

## Verification Goal

1. Successful LLM calls with provider `usage` update the correct Asia/Shanghai 12-hour aggregate window.
2. Window rows survive process restarts because aggregate state is checkpointed in MySQL.
3. Expired aggregate windows are finalized and exposed through a read API.
4. Tests, build, and `go fix` pass after the change.
