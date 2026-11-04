# TASK-01: HTTP Request ID Middleware with Header Tracing (X-Request-ID)

- **Type:** Task (Feature)
- **Status:** Complete
- **Estimated Effort:** 2 hours (<8h)
- **Target Branch:** `feature/01-http-request-id-middleware`
- **Created Date:** 2026-11-04
- **Completed Date:** 2026-11-04 

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
- [x] **AC-1:** `RequestID` middleware inspects incoming requests for an `X-Request-ID` header: if present and non-empty, it preserves and reuses it; if missing or empty, it generates a new UUID v4 string.
- [x] **AC-2:** Outgoing HTTP response always includes the `X-Request-ID` header matching the active request ID.
- [x] **AC-3:** The request ID is injected into the request's `context.Context` and retrievable via an exported helper function `GetRequestID(ctx context.Context) string` (or empty string if not found).
- [x] **AC-4:** `Logging` middleware extracts the request ID from context and includes `"request_id"` in the structured `slog.Info` log entry.
- [x] **AC-5:** `Recovery` middleware extracts the request ID from context and includes `"request_id"` in the panic recovery `slog.Error` log entry.
- [x] **AC-6:** `NewRouter` in `internal/http/router.go` includes `RequestID` in the middleware stack so all registered routes are traced.
- [x] **AC-7:** Accompanying unit tests authored in RED state first under `internal/http/` covering UUID generation, incoming header preservation, context retrieval, and router integration.
- [x] **AC-8:** Full test suite (`go test -v -race ./internal/http/...`) and Go build (`go build ./...`) pass with zero errors.

## 5. DoD Independent Audit Log (`dod-checker`)
*(To be populated by the `dod-checker` subagent upon completion)*
- [x] AC-1: [PASS] `internal/http/middleware.go:31-34` & verified by `internal/http/middleware_test.go:64-71, 116-119`
- [x] AC-2: [PASS] `internal/http/middleware.go:36` & verified by `internal/http/middleware_test.go:59-62`
- [x] AC-3: [PASS] `internal/http/middleware.go:19-27, 37-38` & verified by `internal/http/middleware_test.go:74-76, 128-151`
- [x] AC-4: [PASS] `internal/http/middleware.go:70-75` & verified by `internal/http/middleware_test.go:181-187`
- [x] AC-5: [PASS] `internal/http/middleware.go:86-90` & verified by `internal/http/middleware_test.go:262-268`
- [x] AC-6: [PASS] `internal/http/router.go:41-45` & verified by `internal/http/router_test.go:27-38, 67-70, 90-93, 111-114`
- [x] AC-7: [PASS] `internal/http/middleware_test.go:17-274` and `internal/http/router_test.go:14-115`
- [x] AC-8: [PASS] Verified clean compile and assertions across package and build
