# Go Interview Crawler Service Design

**Goal:** Build a Go HTTP service that crawls interview-experience posts from NowCoder first, stores raw posts with multi-platform dedupe in MySQL, extracts structured interview information with pluggable LLM backends, and exposes job/post/question query APIs.

## Scope

- Support `nowcoder` as the first crawler platform.
- Support future platforms through a crawler interface and `(platform, source_post_id)` dedupe.
- Expose HTTP APIs for health, crawl jobs, posts, companies, questions, and single-post reparse.
- Run scheduled crawl jobs and manual jobs through the same execution pipeline.
- Persist raw posts, parse results, and per-question rows for later knowledge graph/tag analysis.

## Non-Goals

- No frontend in this task.
- No MQ-based distributed execution in this version.
- No advanced anti-bot or browser-based scraping in this first pass.
- No provider-specific LLM integrations beyond OpenAI-compatible transport in the initial implementation.

## Architecture

The service is an asynchronous monolith. HTTP handlers and the scheduler only create jobs. A background worker pulls jobs, calls platform crawlers, deduplicates raw posts, and sends new or updated posts to the LLM extraction pipeline. Query APIs read directly from MySQL.

## Main Components

### HTTP API

- `GET /health`
- `POST /api/v1/crawl/jobs`
- `GET /api/v1/crawl/jobs`
- `GET /api/v1/crawl/jobs/:id`
- `GET /api/v1/posts`
- `GET /api/v1/posts/:platform/:sourcePostID`
- `POST /api/v1/posts/:platform/:sourcePostID/reparse`
- `GET /api/v1/questions`
- `GET /api/v1/companies`

### Job Executor

- Receives manual and scheduled jobs.
- Prevents overlapping equivalent jobs with an in-process lock.
- Records job lifecycle and per-post disposition.
- Retries platform page fetches and LLM extraction with bounded backoff.

### Platform Crawler Interface

```go
type PlatformCrawler interface {
    Platform() string
    SearchPosts(ctx context.Context, req SearchRequest) ([]RawPostInput, error)
}
```

- First implementation: `NowCoderCrawler`
- Future implementations: `ZhihuCrawler`, `XiaohongshuCrawler`, others

### LLM Extractor Interface

```go
type LLMExtractor interface {
    Name() string
    Extract(ctx context.Context, post RawPostForLLM) (*StructuredPost, error)
}
```

- First implementation: `OpenAICompatibleExtractor`
- Future implementations: `CodexCLIExtractor`, `ClaudeSDKExtractor`, `OpenCodeExtractor`

## Storage Model

### `crawl_jobs`

- Tracks one crawl execution.
- Important fields: `id`, `trigger_type`, `status`, `platforms_json`, `keywords_json`, `pages`, `force_reparse`, `started_at`, `finished_at`, `stats_json`, `error_message`

### `raw_posts`

- Main raw-storage and dedupe table.
- Important fields: `id`, `platform`, `source_post_id`, `title`, `content`, `content_hash`, `author_name`, `post_url`, `company_name_raw`, `company_name_norm`, `source_created_at`, `source_edited_at`, `last_crawled_at`, `parse_status`, `parse_attempts`, `raw_payload_json`
- Unique key: `(platform, source_post_id)`

### `post_parse_results`

- One current structured record per raw post.
- Important fields: `raw_post_id`, `schema_version`, `llm_provider`, `llm_model`, `company_name`, `sentiment`, `sentiment_reason`, `key_events_json`, `questions_json`, `tags_json`, `raw_json`, `parsed_at`

### `interview_questions`

- One interview question per row under the minimal-line rule.
- Important fields: `id`, `raw_post_id`, `platform`, `company_name`, `question_text`, `question_order`, `category`, `tags_json`, `source_excerpt`, `created_at`

### `question_tags`

- Optional normalized tag table for graph-style exploration.
- Important fields: `question_id`, `tag`

### `crawl_job_posts`

- Auditable mapping from jobs to processed posts and disposition.
- Important fields: `job_id`, `raw_post_id`, `disposition`, `message`

## Dedupe and Update Rules

- First key: `(platform, source_post_id)`
- Update decision: compare `source_edited_at`
- Fallback update detection: compare `content_hash`
- Processing result:
  - new post -> insert and parse
  - unchanged post -> mark unchanged, no parse
  - changed post -> update raw row and reparse

## Structured JSON Schema

```json
{
  "schema_version": "v1",
  "company": {
    "raw_name": "阿里淘天",
    "normalized_name": "阿里巴巴",
    "confidence": 0.92
  },
  "post_type": "interview_experience",
  "sentiment": {
    "label": "positive",
    "confidence": 0.81,
    "reason": "整体反馈顺利，问题偏基础"
  },
  "key_events": [
    {
      "type": "interview_round",
      "summary": "一面偏后端基础和消息队列",
      "round": "first"
    }
  ],
  "questions": [
    {
      "order": 1,
      "question": "Kafka 为什么能保证高吞吐？",
      "category": "backend",
      "tags": ["kafka", "MQ", "后端", "阿里"],
      "source_excerpt": "问了 Kafka 为什么吞吐高"
    }
  ],
  "overall_tags": ["阿里", "淘天", "后端", "校招"]
}
```

## Minimal-Line Rule

- Every distinct interview question becomes exactly one row.
- Parallel questions in the same sentence must be split.
- Each row must contain `question`, `tags`, `category`, and `source_excerpt`.
- Tags may mix technical, domain, and company vocabulary in the first version.

## Error Handling

- Platform fetch errors retry per page with bounded backoff.
- LLM failures set `parse_status=failed` and preserve raw responses for later reparse.
- Low-confidence company recognition may keep `company_name_norm` empty rather than guessing.

## Testing Strategy

- Unit tests for request parsing, dedupe decisions, JSON schema handling, and question flattening.
- Integration tests with mocked NowCoder API and mocked LLM extractor.
- Repository tests should verify repeated crawl behavior, updated post reparsing, and query API outputs.
