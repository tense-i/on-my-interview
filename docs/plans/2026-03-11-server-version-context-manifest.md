## Task

Add a fixed server version string in `server/cmd/server/main.go`, print it on startup, and expose it through `/health`.

## Loaded Context

- `AGENTS.md`
  - Repo workflow requirements and context-manifest rule.
- `rules/10_release.md`
  - Release discipline and tag/version expectations.
- `server/cmd/server/main.go`
  - Current boot path where startup logging should occur.
- `server/internal/app/app.go`
  - App construction path where version must be passed to the router.
- `server/internal/http/router.go`
  - Router construction point for wiring health dependencies.
- `server/internal/http/router_test.go`
  - Existing health route test to extend.
- `server/internal/http/handlers/health.go`
  - Current health response shape.

## Excluded Context

- Storage, crawler, and LLM implementation files
  - Not relevant to version string exposure.
- Frontend or non-server directories
  - Out of scope.

## Assumptions

- The version string should be a fixed code constant in `main.go`.
- The same version must appear in startup logs and `/health`.
- Version changes will be made manually during release preparation.

## Verification Goal

1. Startup logs include the configured version string.
2. `/health` returns both `status` and `version`.
3. Tests and build continue to pass.
