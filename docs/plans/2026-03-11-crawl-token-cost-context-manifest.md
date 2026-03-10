## Task

Run another bounded NowCoder crawl with real DeepSeek parsing, persist additional posts, and analyze average token usage and estimated cost for the newly parsed batch.

## Loaded Context

- `AGENTS.md`
  - Context manifest requirement and verification-first workflow.
- `memory/FACTS.md`
  - Current runtime commands, DeepSeek-compatible base URL, storage schema.
- `memory/EXPERIENCE.md`
  - DeepSeek schema-prompt pitfall and NowCoder empty-record filtering pitfall.
- `server/internal/llm/openai/extractor.go`
  - Exact request path and system prompt used by the service.
- `server/internal/http/router.go`
  - Trigger and query endpoints for crawl jobs and result inspection.
- DeepSeek official docs:
  - `https://api-docs.deepseek.com/zh-cn/quick_start/pricing`
  - `https://api-docs.deepseek.com/zh-cn/quick_start/token_usage`

## Excluded Context

- Frontend-related files and docs
  - Not relevant to crawler execution or token accounting.
- Most implementation docs from earlier feature work
  - Useful background, but not required for this operational run.
- Any persistent code changes for token accounting
  - Prefer temporary runtime instrumentation first.

## Assumptions

- Local Docker MySQL and Adminer containers remain healthy.
- The previously provided DeepSeek key is still valid for this runtime verification.
- Token cost analysis should cover only the newly parsed posts from this batch.

## Verification Goal

1. Insert and parse several additional posts into MySQL.
2. Capture per-request DeepSeek `usage` from real parsing traffic.
3. Compute average prompt tokens, completion tokens, total tokens, and estimated average cost per parsed post using current official pricing.
