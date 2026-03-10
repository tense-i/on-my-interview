# Server Version Exposure Design

**Goal:** Add a fixed server version string in the Go server entrypoint, print it during startup, and expose it through `/health`.

## Scope

- Define a fixed version constant in `server/cmd/server/main.go`
- Print the version when the process starts
- Pass the version into the HTTP layer
- Return the version in `/health`
- Update tests and README if needed

## Non-Goals

- Do not add build-time injection or environment-based versioning
- Do not add a dedicated `/version` endpoint
- Do not propagate the version into database records or other APIs

## Recommended Approach

Use a package-level constant in `main.go`, set to the current release version string. `main` prints this version before building the app, then passes it into `app.New`. The app forwards the string to the router, and the health handler returns it alongside the existing `status`.

This keeps the change minimal and explicit. Version changes remain a deliberate source edit during release prep, matching the requested workflow.

## Code Changes

- `server/cmd/server/main.go`
  - add `const version = "v0.0.1"`
  - print startup log with the version
  - pass version into `app.New`
- `server/internal/app/app.go`
  - extend `New` to accept a version string
  - pass version into router dependencies
- `server/internal/http/router.go`
  - extend dependencies to carry version
  - wire health handler with version
- `server/internal/http/handlers/health.go`
  - change health handler to include version in JSON
- `server/internal/http/router_test.go`
  - assert `/health` returns `status` and `version`

## API Shape

Before:

```json
{"status":"ok"}
```

After:

```json
{"status":"ok","version":"v0.0.1"}
```

## Risks

- Hard-coded version requires manual edit for each release
- Tests that rely on the exact `/health` payload must be updated

## Verification

- `cd server && go test ./internal/http -count=1`
- `cd server && go test ./...`
- `cd server && go build -buildvcs=false ./...`
- Run the binary and confirm startup log contains the version
