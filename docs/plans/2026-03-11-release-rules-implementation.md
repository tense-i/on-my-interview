# Release Rules Update Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Update the repository release rules for the Go server workflow, then commit the pending release-related changes and move/push `v0.0.1`.

**Architecture:** The release rules will be rewritten around the fixed version constant in `server/cmd/server/main.go`. After the docs update, the pending version feature files and new planning docs will be committed, and the local/remote `v0.0.1` tag will be moved to that release commit.

**Tech Stack:** Markdown docs, Git tags, Go verification commands.

---

### Task 1: Rewrite release rules for the Go workflow

**Files:**
- Modify: `rules/10_release.md`

**Step 1: Review current mismatch**

Confirm the file still references `package.json` and `npm install`.

**Step 2: Write minimal replacement**

Update the file to describe:
- changing `server/cmd/server/main.go` version
- running Go verification commands
- committing the version bump
- tagging and pushing only when explicitly instructed

**Step 3: Review the resulting rules**

Run: `sed -n '1,240p' rules/10_release.md`
Expected: file reflects the Go/server release flow only.

### Task 2: Verify pending release changes

**Files:**
- Verify: `server/cmd/server/main.go`
- Verify: `server/internal/app/app.go`
- Verify: `server/internal/http/router.go`
- Verify: `server/internal/http/handlers/health.go`
- Verify: `server/internal/http/router_test.go`
- Verify: `docs/plans/2026-03-11-server-version-*.md`

**Step 1: Run verification suite**

Run:
- `cd server && GOTOOLCHAIN=go1.26.0 go fix -diff ./...`
- `cd server && GOTOOLCHAIN=go1.26.0 go fix ./...`
- `cd server && go test ./...`
- `cd server && go build -buildvcs=false ./...`

**Step 2: Confirm working tree scope**

Run: `git status --short`
Expected: only intended release-related files remain modified/untracked.

### Task 3: Commit and move the release tag

**Files:**
- Commit: release rules + pending version feature files + new planning docs

**Step 1: Stage intended files**

Stage:
- `rules/10_release.md`
- pending server version files
- associated plan docs

**Step 2: Create release-focused commit**

Use a conventional commit summarizing:
- version exposure
- release rule alignment

**Step 3: Move local tag**

Run:
- `git tag -fa v0.0.1 -m "v0.0.1"`

Expected: local tag now points to the new commit.

**Step 4: Push commit and force-update tag**

Run:
- `git push origin main`
- `git push --force origin v0.0.1`

Expected: remote branch and tag both point to the new release commit.
