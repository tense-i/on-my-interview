# Release Rules

## Release Flow
1. Update `package.json` version → run `npm install` → commit `chore: bump version to X.Y.Z`
2. Push → `git tag vX.Y.Z` → `git push origin vX.Y.Z`
3. CI auto-builds and publishes GitHub Release

## Release Discipline
- Never auto-release: do not run `git push`, `git tag`, or `git push origin vX.Y.Z` unless user explicitly instructs
- If ambiguous, ask before any push/tag

## Release Notes Sections
1. `更新内容`
2. `Downloads`
3. `Installation`
4. `Requirements`
5. `Changelog`

## CI Triggers
- `.github/workflows/ci.yml` - tags `v*`

# Commit Rules

## Commit Convention
- Title line: use conventional commits format (feat/fix/refactor/chore, etc.)
- Body: group by file or feature, explain what changed, why, and impact scope
- Bug fixes: state root cause; architecture decisions: briefly explain rationale


