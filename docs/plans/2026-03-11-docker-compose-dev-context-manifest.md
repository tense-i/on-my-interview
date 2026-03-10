# Context Manifest: Docker Compose Development Environment

**Task:** Add a local Docker Compose development environment for the Go interview crawler service, using Compose for MySQL and Adminer while the Go service continues to run on the host machine.

**Date:** 2026-03-11

## Loaded Context

- `AGENTS.md`
  - Reason: repository workflow, mandatory skills, verification, and memory update rules.
- `rules/00_core.md`
  - Reason: context discipline, manifest requirement, and output contract.
- `rules/10_release.md`
  - Reason: release and commit guardrails.
- `memory/FACTS.md`
  - Reason: confirm the current Go service layout and runtime surface.
- `memory/EXPERIENCE.md`
  - Reason: check reusable pitfalls before changing local dev tooling.
- `server/README.md`
  - Reason: inspect current local run instructions and environment-variable expectations.
- `server/internal/app/app.go`
  - Reason: confirm runtime dependencies and startup behavior for MySQL and scheduler wiring.
- `server/internal/config/config.go`
  - Reason: inspect environment loading behavior and current config surface.
- `docs/plans/2026-03-10-nowcoder-crawler-context-manifest.md`
  - Reason: reuse recent backend context and avoid reloading unrelated code.
- `docs/plans/2026-03-10-nowcoder-crawler-design.md`
  - Reason: preserve consistency with the already approved backend design.
- `/Users/zh/.codex/skills/brainstorming/SKILL.md`
  - Reason: required feature-design workflow.
- `/Users/zh/.codex/skills/writing-plans/SKILL.md`
  - Reason: required pre-implementation planning workflow.

## Context Excluded

- Most files under `newcoder-crawler/`
  - Reason: the Docker Compose task only affects local development for the Go service, not the legacy Python prototype.
- Query handlers, crawler logic, and storage details not related to startup/config
  - Reason: the task is infrastructure-focused and should keep code-context narrow.
- UI/browser verification skills
  - Reason: this task adds local infrastructure and backend config only, with no frontend changes.

## Observed Constraints

- The repository currently has no `docker-compose*.yml`, `Dockerfile`, or `.env.example`.
- The Go service already auto-applies MySQL migrations on startup.
- The Go service currently relies on process environment variables only; it does not load a `.env` file.
- The desired dev shape is host-run Go service plus Compose-managed `mysql` and `adminer`.

## Assumptions Chosen

- Use a root-level `docker-compose.dev.yml` for local-only infrastructure.
- Use a root-level `.env.example` shared by Compose and host-run Go development.
- Add lightweight `.env` loading in Go so `go run ./cmd/server` can work directly with a root `.env`.
- Keep production/runtime behavior unchanged when environment variables are already set externally.
