## Task

Update the repository release rules to match the Go server release workflow, then commit the pending release-related changes and move/push tag `v0.0.1` to the new release commit.

## Loaded Context

- `AGENTS.md`
  - Repo workflow, context manifest requirement, release discipline.
- `rules/10_release.md`
  - Current release rules, which still describe an npm/package.json flow.
- `server/cmd/server/main.go`
  - Fixed version string now used by startup logging and `/health`.
- `docs/plans/2026-03-11-server-version-*.md`
  - Recently documented server version change that should be reflected in release rules.
- `git status`, `git tag`, `git show v0.0.1`
  - Current working tree and tag position.

## Excluded Context

- Crawler/storage/LLM implementation files
  - Not relevant to release-rule wording.
- Frontend or unused directories
  - Out of scope.

## Assumptions

- The release process should now be defined around the Go server, not npm.
- The version source of truth is the fixed `version` constant in `server/cmd/server/main.go`.
- The user explicitly wants `v0.0.1` to be moved and pushed after the new commit is created.

## Verification Goal

1. `rules/10_release.md` documents the Go/server release flow accurately.
2. The pending release-related code/docs changes are committed.
3. Local tag `v0.0.1` points to the new release commit.
4. Remote tag `v0.0.1` is force-updated to the same commit.
