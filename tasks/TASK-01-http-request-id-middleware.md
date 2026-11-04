# TASK-01: HTTP Request ID Middleware with Header Tracing (X-Request-ID)

- **Type:** Task (Feature)
- **Status:** In Progress
- **Estimated Effort:** 2 hours (<8h)
- **Target Branch:** `feature/01-http-request-id-middleware`
- **Created Date:** 2026-11-04
- **Completed Date:** 

## 1. Context & User Story
As a backend engineer and system operator, I want all incoming HTTP requests to be tracked with a unique `X-Request-ID` header and context identifier, so that individual HTTP requests can be traced, correlated with structured application logs (`slog`), and easily debugged across system boundaries.

## 2. Problem Statement / Requirement
`internal/http` currently provides `Logging` and `Recovery` middleware, but neither injects nor preserves request tracing identifiers. Downstream handlers and logging statements lack correlation IDs, and clients cannot track request lifecycles via the standard `X-Request-ID` HTTP header.

## 3. Pre-Flight Codebase Audit
- **Files Inspected:**
  - `internal/http/middleware.go`
  - `internal/http/router.go`
  - `internal/http/slack.go`
  - `cmd/server/main.go`
  - `go.mod`
- **Prior Art / Reusable Logic:**
  - `internal/http/middleware.go` already implements standard `http.Handler` middleware wrappers (`Logging`, `Recovery`) and a custom `responseWriter`.
  - `github.com/google/uuid` (v1.6.0) is already available in `go.mod`.
  - Structured logging is standardized on standard library `log/slog`.
  - Confirmed net-new middleware functionality; no prior request ID logic exists in the codebase.

## 4. Acceptance Criteria (DoD Contract)
- [ ] **AC-1:** `RequestID` middleware inspects incoming requests for an `X-Request-ID` header: if present and non-empty, it preserves and reuses it; if missing or empty, it generates a new UUID v4 string.
- [ ] **AC-2:** Outgoing HTTP response always includes the `X-Request-ID` header matching the active request ID.
- [ ] **AC-3:** The request ID is injected into the request's `context.Context` and retrievable via an exported helper function `GetRequestID(ctx context.Context) string` (or empty string if not found).
- [ ] **AC-4:** `Logging` middleware extracts the request ID from context and includes `"request_id"` in the structured `slog.Info` log entry.
- [ ] **AC-5:** `Recovery` middleware extracts the request ID from context and includes `"request_id"` in the panic recovery `slog.Error` log entry.
- [ ] **AC-6:** `NewRouter` in `internal/http/router.go` includes `RequestID` in the middleware stack so all registered routes are traced.
- [ ] **AC-7:** Accompanying unit tests authored in RED state first under `internal/http/` covering UUID generation, incoming header preservation, context retrieval, and router integration.
- [ ] **AC-8:** Full test suite (`go test -v -race ./internal/http/...`) and Go build (`go build ./...`) pass with zero errors.

## 5. DoD Independent Audit Log (`dod-checker`)
*(To be populated by the `dod-checker` subagent upon completion)*
- [ ] AC-1: [PENDING]
- [ ] AC-2: [PENDING]
- [ ] AC-3: [PENDING]
- [ ] AC-4: [PENDING]
- [ ] AC-5: [PENDING]
- [ ] AC-6: [PENDING]
- [ ] AC-7: [PENDING]
- [ ] AC-8: [PENDING]
