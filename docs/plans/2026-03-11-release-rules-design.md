# Release Rules Update Design

**Goal:** Replace the outdated npm-based release rules with a Go/server release flow that matches the repository's current version source of truth and release mechanics.

## Scope

- Update `rules/10_release.md`
- Keep the rule set focused on this repository's current Go service release process
- Commit the pending version-exposure changes together with the rules update
- Move and push tag `v0.0.1` to the resulting release commit

## Non-Goals

- Do not redesign CI workflows
- Do not change the tag naming scheme
- Do not introduce build-time version injection

## Recommended Approach

Document the release process around the existing fixed version constant in `server/cmd/server/main.go`.

Release flow should explicitly say:

1. update `server/cmd/server/main.go` version string
2. run required verification:
   - `GOTOOLCHAIN=go1.26.0 go fix -diff ./...`
   - `GOTOOLCHAIN=go1.26.0 go fix ./...`
   - `go test ./...`
   - `go build -buildvcs=false ./...`
3. commit the release version change
4. tag `vX.Y.Z`
5. push commit and tag when explicitly instructed

This keeps the rules aligned with the codebase as it exists today, where the version is manually changed in source for each release.

## Risks

- Because version is source-controlled, release prep must remember to update `main.go`
- Moving an existing tag requires explicit force-push semantics, which must remain user-directed

## Verification

- Review `rules/10_release.md` for consistency with the actual Go release path
- Confirm current server version remains `v0.0.1`
- Verify local tag `v0.0.1` points to the new commit
- Verify remote tag update succeeds
