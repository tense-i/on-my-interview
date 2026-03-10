# Docker Compose Development Environment Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Docker Compose local development environment for MySQL and Adminer, plus root `.env` support so the host-run Go service can start without manual environment exports.

**Architecture:** The Go service remains a host process. `docker-compose.dev.yml` provides only local infrastructure. A lightweight `.env` loader in the Go config package reads the root `.env` before config evaluation, while still allowing explicit environment variables to override file values.

**Tech Stack:** Docker Compose, MySQL 8.4, Adminer 4, Go 1.24.6, standard-library file parsing.

---

### Task 1: Add `.env` loading to the Go config package

**Files:**
- Modify: `server/internal/config/config.go`
- Create: `server/internal/config/config_test.go`

**Step 1: Write the failing test**

Add tests that verify:
- a root `.env` file is loaded when the related environment variable is unset
- an already-set environment variable is not overwritten by the `.env` file

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/config -run TestLoadFromEnv -v`
Expected: FAIL because `.env` loading does not exist yet.

**Step 3: Write minimal implementation**

- Load the repository root `.env` from `../.env` relative to `server/`
- Parse simple `KEY=VALUE` lines
- Ignore blank lines and comments
- Only call `os.Setenv` for keys not already present

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/config -run TestLoadFromEnv -v`
Expected: PASS

### Task 2: Add Docker Compose infrastructure files

**Files:**
- Create: `docker-compose.dev.yml`
- Create: `.env.example`

**Step 1: Write the failing verification**

Run: `docker compose -f docker-compose.dev.yml config`
Expected: FAIL because the Compose file does not exist yet.

**Step 2: Write minimal implementation**

- Define `mysql` with health check, volume, environment variables, and port mapping
- Define `adminer` with dependency on `mysql` and exposed port
- Add matching `.env.example` values for Compose and host-run Go config

**Step 3: Run verification to confirm it parses**

Run: `docker compose -f docker-compose.dev.yml config`
Expected: PASS and rendered config output

### Task 3: Document the local workflow

**Files:**
- Modify: `server/README.md`
- Modify: `memory/FACTS.md`
- Modify: `memory/EXPERIENCE.md` only if a reusable dev-environment pitfall appears during verification

**Step 1: Write the failing verification**

Review the current README and confirm it does not describe the Compose-based local workflow.

**Step 2: Write minimal implementation**

- Add a “Local Dev with Docker Compose” section
- Document `.env` setup, Compose startup, host-run server startup, Adminer access, and health endpoint check
- Update long-lived facts if the workflow becomes a stable project fact

**Step 3: Run verification**

Run: `sed -n '1,260p' server/README.md`
Expected: README now contains the new local workflow section

### Task 4: Full verification

**Files:**
- Verify: `docker-compose.dev.yml`
- Verify: `.env.example`
- Verify: `server/internal/config/config.go`
- Verify: `server/internal/config/config_test.go`
- Verify: `server/README.md`

**Step 1: Run the verification suite**

Run:
- `cd server && go test ./...`
- `cd server && go build -buildvcs=false ./...`
- `docker compose -f docker-compose.dev.yml config`

**Step 2: Run infrastructure smoke test**

Run:
- `docker compose -f docker-compose.dev.yml up -d`
- `docker compose -f docker-compose.dev.yml ps`

**Step 3: Clean up if needed**

Run:
- `docker compose -f docker-compose.dev.yml down`

**Step 4: Confirm final requirements**

Verify the implementation includes:
- Compose-managed MySQL and Adminer
- Root `.env.example` with host-run and Compose variables
- Host-run Go service `.env` loading support
- Updated local dev documentation
