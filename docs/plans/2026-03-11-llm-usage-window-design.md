# LLM Usage Window Aggregation Design

**Goal:** Persist global LLM token usage and estimated cost into fixed 12-hour Asia/Shanghai windows, without storing per-request detail, and expose a factual query API over finalized windows only.

## Scope

- Capture provider `usage` from successful LLM responses.
- Aggregate usage and cost into 12-hour windows keyed by Asia/Shanghai natural half-days.
- Persist aggregate checkpoints in MySQL so service restarts do not lose current-window totals.
- Finalize expired windows as part of normal write/read flows.
- Expose an HTTP API to list finalized usage windows and range totals.

## Non-Goals

- Do not store per-request or per-post billing detail.
- Do not expose in-progress window totals through the API.
- Do not add multi-provider billing logic in this task.
- Do not add dashboards or external observability systems.

## Requirements

- Aggregation granularity is fixed to Beijing time natural windows:
  - `00:00:00` to `12:00:00`
  - `12:00:00` to `24:00:00`
- Aggregation is global, not grouped by provider/model/platform in the API.
- The API must return factual settled data only.
- The service must not lose partial-window totals on restart.

## Recommended Approach

Use a single MySQL-backed aggregate window table that stores both:

- the current checkpointed aggregate for an in-progress 12-hour window, and
- finalized historical records for expired windows.

Each successful LLM response with `usage` updates the current aggregate row via an atomic UPSERT-like flow. Before updating the current row, the repository marks any expired `aggregating` rows as `finalized`. This preserves restart safety without storing request-level rows.

## Data Model

Add one table: `llm_usage_windows`.

Suggested columns:

- `id BIGINT PRIMARY KEY AUTO_INCREMENT`
- `window_start_at DATETIME NOT NULL`
- `window_end_at DATETIME NOT NULL`
- `timezone VARCHAR(64) NOT NULL`
- `status VARCHAR(32) NOT NULL`
  - `aggregating`
  - `finalized`
- `request_count BIGINT NOT NULL`
- `usage_observed_count BIGINT NOT NULL`
- `prompt_tokens BIGINT NOT NULL`
- `prompt_cache_hit_tokens BIGINT NOT NULL`
- `prompt_cache_miss_tokens BIGINT NOT NULL`
- `completion_tokens BIGINT NOT NULL`
- `total_tokens BIGINT NOT NULL`
- `estimated_cost_cny DECIMAL(20,8) NOT NULL`
- `created_at DATETIME NOT NULL`
- `updated_at DATETIME NOT NULL`
- `finalized_at DATETIME NULL`

Constraints:

- unique key on `window_start_at`

This keeps one row per Beijing half-day and avoids separate checkpoint/history tables.

## Usage Capture

The OpenAI-compatible extractor should decode the provider `usage` object in addition to the structured content payload. DeepSeek documents that chat completion responses include:

- `prompt_tokens`
- `prompt_cache_hit_tokens`
- `prompt_cache_miss_tokens`
- `completion_tokens`
- `total_tokens`

The service should only attempt billing aggregation when `usage` is present.

## Cost Calculation

The first cost calculator targets DeepSeek `deepseek-chat`.

Pricing source:
- DeepSeek official pricing page currently states:
  - cache hit input: `0.2` CNY / 1M tokens
  - cache miss input: `2` CNY / 1M tokens
  - output: `3` CNY / 1M tokens

Calculation:

- `cost_cny = prompt_cache_hit_tokens * 0.2 / 1_000_000`
- `+ prompt_cache_miss_tokens * 2 / 1_000_000`
- `+ completion_tokens * 3 / 1_000_000`

Store the result as `DECIMAL(20,8)` and round in Go before persistence.

If the provider returns `usage` but one of the cache fields is absent, treat the missing field as zero and preserve the reported `prompt_tokens`/`total_tokens`.

## Window Semantics

Window selection is based on `time.LoadLocation("Asia/Shanghai")`.

Rules:

- if local hour `< 12`, the window starts at local `00:00:00`
- otherwise the window starts at local `12:00:00`
- `window_end_at` is exactly 12 hours after start

Finalization:

- before any aggregate update, finalize all rows where:
  - `status = 'aggregating'`
  - `window_end_at <= now`
- also finalize expired rows on read requests to avoid a stale API when traffic resumes after an idle period

## API Design

Add:

- `GET /api/v1/usage/windows`

Query params:

- `limit` default `50`
- `offset` default `0`
- `from` optional RFC3339 lower bound on `window_start_at`
- `to` optional RFC3339 upper bound on `window_end_at`

Response shape:

```json
{
  "items": [
    {
      "window_start_at": "2026-03-11T00:00:00+08:00",
      "window_end_at": "2026-03-11T12:00:00+08:00",
      "timezone": "Asia/Shanghai",
      "request_count": 42,
      "usage_observed_count": 42,
      "prompt_tokens": 12345,
      "prompt_cache_hit_tokens": 4000,
      "prompt_cache_miss_tokens": 8345,
      "completion_tokens": 6789,
      "total_tokens": 19134,
      "estimated_cost_cny": "0.04123456",
      "finalized_at": "2026-03-11T12:00:01+08:00"
    }
  ],
  "totals": {
    "request_count": 42,
    "usage_observed_count": 42,
    "prompt_tokens": 12345,
    "prompt_cache_hit_tokens": 4000,
    "prompt_cache_miss_tokens": 8345,
    "completion_tokens": 6789,
    "total_tokens": 19134,
    "estimated_cost_cny": "0.04123456"
  }
}
```

Only `finalized` rows are returned.

## Code Placement

- `server/internal/llm`
  - extend structured extraction result to carry provider usage
- `server/internal/llm/openai`
  - decode `usage` from the provider response
- `server/internal/jobs`
  - pass usage into the recorder only on successful extractions
- `server/internal/storage/repository`
  - define usage window models and repository methods
- `server/internal/storage/mysql`
  - implement aggregate write/finalize/query logic
- `server/internal/http/handlers`
  - add usage window list endpoint
- `server/internal/http/router.go`
  - register the new route

## Failure Handling

- LLM request fails:
  - no aggregate update
- LLM request succeeds but `usage` is missing:
  - aggregate `request_count` only if we explicitly choose to count all successful requests
  - for this task, count only requests with observed usage to keep billing rows factual and simple
- Aggregate update fails after a successful parse:
  - return an error from the job path so the issue is visible and not silently ignored

## Testing Strategy

- unit tests for:
  - Asia/Shanghai 12-hour window selection
  - DeepSeek cost calculation
  - extractor decoding of `usage`
- repository tests for:
  - aggregating into one window across multiple writes
  - finalizing expired windows
  - listing finalized windows with totals
- HTTP tests for:
  - `GET /api/v1/usage/windows`
- integration verification:
  - run a real crawl job
  - confirm rows land in `llm_usage_windows`
  - query API returns finalized windows once the repository is seeded appropriately in tests
