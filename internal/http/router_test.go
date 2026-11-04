package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	apphttp "github.com/luiggisao/auto-golpe/internal/http"
)

func TestNewRouter_HealthRoute_AutoGeneratesRequestID(t *testing.T) {
	// AC-6: NewRouter wraps handlers with RequestID middleware so /health returns X-Request-ID
	router := apphttp.NewRouter(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK from /health, got %d", rec.Code)
	}

	reqID := rec.Header().Get("X-Request-ID")
	if reqID == "" {
		t.Fatalf("expected /health to return non-empty X-Request-ID header, got empty")
	}

	parsed, err := uuid.Parse(reqID)
	if err != nil {
		t.Fatalf("/health X-Request-ID %q is not a valid UUID: %v", reqID, err)
	}
	if parsed.Version() != 4 {
		t.Fatalf("expected /health X-Request-ID UUID version 4, got %d", parsed.Version())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode /health JSON body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %v", body["status"])
	}
	if body["app"] != "golpe" {
		t.Fatalf("expected app 'golpe', got %v", body["app"])
	}
}

func TestNewRouter_HealthRoute_PreservesIncomingRequestID(t *testing.T) {
	// AC-6: NewRouter preserves incoming X-Request-ID on /health
	router := apphttp.NewRouter(nil, nil)

	customID := "incoming-custom-health-trace-999"
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", customID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	gotID := rec.Header().Get("X-Request-ID")
	if gotID != customID {
		t.Fatalf("expected /health to preserve X-Request-ID %q, got %q", customID, gotID)
	}
}

func TestNewRouter_EchoRoute_HasRequestID(t *testing.T) {
	// AC-6: All registered routes in NewRouter include X-Request-ID
	router := apphttp.NewRouter(nil, nil)

	bodyJSON := `{"message":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/echo", strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	customID := "echo-trace-id-abc"
	req.Header.Set("X-Request-ID", customID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK from /api/v1/echo, got %d", rec.Code)
	}

	gotID := rec.Header().Get("X-Request-ID")
	if gotID != customID {
		t.Fatalf("expected /api/v1/echo to preserve X-Request-ID %q, got %q", customID, gotID)
	}
}

func TestNewRouter_PanicRoute_RecoversAndHasRequestID(t *testing.T) {
	// AC-6: Panic recovery through NewRouter still returns X-Request-ID
	router := apphttp.NewRouter(nil, nil)

	customID := "panic-trace-id-def"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil)
	req.Header.Set("X-Request-ID", customID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 InternalServerError from /api/v1/panic, got %d", rec.Code)
	}

	gotID := rec.Header().Get("X-Request-ID")
	if gotID != customID {
		t.Fatalf("expected /api/v1/panic response to preserve X-Request-ID %q, got %q", customID, gotID)
	}
}
