## Task

Validate the existing crawler service end-to-end against the real DeepSeek OpenAI-compatible API without changing repository configuration files or persisting secrets.

## Loaded Context

- `AGENTS.md`
  - Repo workflow requirements, context manifest requirement, verification expectations.
- `memory/FACTS.md`
  - Existing project layout and runtime commands.
- `memory/EXPERIENCE.md`
  - Prior MySQL nullable scan pitfall and repo-specific runtime lessons.
- `server/internal/config/config.go`
  - Environment variable surface and `.env` loading behavior.
- `server/README.md`
  - Local startup flow and API endpoints.
- `server/cmd/server/main.go`
  - Service boot entrypoint.
- `server/internal/http/router.go`
  - Available HTTP endpoints used for verification.

## Excluded Context

- Frontend or unrelated docs under `docs/`
  - Not relevant to backend runtime verification.
- Most crawler/parser implementation files
  - Deferred unless runtime verification exposes a concrete defect.
- Existing design and implementation docs for earlier feature work
  - Useful background, but not needed to execute this validation pass.

## Assumptions

- Local Docker daemon is running and the MySQL container from `docker-compose.dev.yml` is healthy.
- The previously supplied DeepSeek API key is intended for this live verification only.
- The API base URL should be `https://api.deepseek.com` and model alias should be `deepseek-chat`.

## Verification Goal

Prove or disprove that the current service can:

1. boot against local MySQL with real DeepSeek credentials,
2. complete one NowCoder crawl job,
3. persist parse results and extracted questions,
4. expose the resulting data through existing query APIs.
