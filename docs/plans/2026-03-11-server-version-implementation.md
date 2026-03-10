# Server Version Exposure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a fixed version string to the Go server, log it at startup, and expose it from `/health`.

**Architecture:** A single source-of-truth version constant lives in `server/cmd/server/main.go`. `main` passes it into `app.New`, the app forwards it to the router, and the health handler returns it in the JSON payload.

**Tech Stack:** Go 1.24.6, Chi router, standard library logging and JSON encoding.

---

### Task 1: Extend health route tests for version output

**Files:**
- Modify: `server/internal/http/router_test.go`

**Step 1: Write the failing test**

Add assertions that `/health` returns both:
- `status = ok`
- `version = v0.0.1`

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/http -run TestRouterHealth -count=1`
Expected: FAIL because the router and handler do not expose `version` yet.

**Step 3: Write minimal implementation**

- Pass a version into the router
- Return that version from the health handler

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/http -run TestRouterHealth -count=1`
Expected: PASS

### Task 2: Wire version from main through app to health

**Files:**
- Modify: `server/cmd/server/main.go`
- Modify: `server/internal/app/app.go`
- Modify: `server/internal/http/router.go`
- Modify: `server/internal/http/handlers/health.go`

**Step 1: Write the failing verification**

Build or test after changing call sites only far enough to surface missing parameters.

Run: `cd server && go test ./...`
Expected: FAIL until the new constructor signature is wired through.

**Step 2: Write minimal implementation**

- Add `const version = "v0.0.1"` in `main.go`
- Log startup version
- Change `app.New` to accept the version
- Pass version into router dependencies
- Return `status` and `version` in health JSON

**Step 3: Run verification to confirm it passes**

Run: `cd server && go test ./...`
Expected: PASS

### Task 3: Final verification

**Files:**
- Verify: `server/cmd/server/main.go`
- Verify: `server/internal/app/app.go`
- Verify: `server/internal/http/router.go`
- Verify: `server/internal/http/handlers/health.go`
- Verify: `server/internal/http/router_test.go`

**Step 1: Run required Go 1.26 fix pass**

Run:
- `cd server && GOTOOLCHAIN=go1.26.0 go fix -diff ./...`
- `cd server && GOTOOLCHAIN=go1.26.0 go fix ./...`

**Step 2: Run full suite**

Run:
- `cd server && go test ./...`
- `cd server && go build -buildvcs=false ./...`

**Step 3: Optional runtime smoke check**

Run:
- `cd server && go run ./cmd/server`
- `curl http://localhost:8080/health`

Expected:
- startup log includes `version=v0.0.1`
- `/health` includes `version`
