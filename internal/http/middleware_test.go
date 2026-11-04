package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	apphttp "github.com/luiggisao/auto-golpe/internal/http"
)

func TestRequestID_GeneratedWhenMissingOrWhitespace(t *testing.T) {
	testCases := []struct {
		name          string
		incomingValue *string
	}{
		{
			name:          "AC-1 & AC-2: missing X-Request-ID header generates valid UUID v4",
			incomingValue: nil,
		},
		{
			name:          "AC-1 & AC-2: empty string X-Request-ID header generates valid UUID v4",
			incomingValue: ptr(""),
		},
		{
			name:          "AC-1 & AC-2: whitespace-only spaces X-Request-ID generates valid UUID v4",
			incomingValue: ptr("    "),
		},
		{
			name:          "AC-1 & AC-2: whitespace-only tabs/newlines X-Request-ID generates valid UUID v4",
			incomingValue: ptr(" \t \n "),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedCtxID string
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtxID = apphttp.GetRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			wrapped := apphttp.RequestID(dummyHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.incomingValue != nil {
				req.Header.Set("X-Request-ID", *tc.incomingValue)
			}
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			// AC-2: Outgoing HTTP response always includes X-Request-ID header
			outgoingHeader := rec.Header().Get("X-Request-ID")
			if outgoingHeader == "" {
				t.Fatalf("expected non-empty X-Request-ID response header, got empty")
			}

			// Validate UUID v4 format
			parsedUUID, err := uuid.Parse(outgoingHeader)
			if err != nil {
				t.Fatalf("outgoing X-Request-ID %q is not a valid UUID: %v", outgoingHeader, err)
			}
			if parsedUUID.Version() != 4 {
				t.Fatalf("expected UUID version 4, got version %d", parsedUUID.Version())
			}

			// AC-3: Context request ID matches outgoing response header
			if capturedCtxID != outgoingHeader {
				t.Fatalf("expected context request ID %q to match response header %q", capturedCtxID, outgoingHeader)
			}
		})
	}
}

func TestRequestID_PreservesExistingHeader(t *testing.T) {
	testCases := []struct {
		name        string
		incomingID  string
	}{
		{
			name:       "AC-1 & AC-2: preserves custom alphanumeric request ID",
			incomingID: "custom-trace-id-abc-12345",
		},
		{
			name:       "AC-1 & AC-2: preserves incoming UUID v4",
			incomingID: uuid.New().String(),
		},
		{
			name:       "AC-1 & AC-2: preserves request ID with special punctuation",
			incomingID: "trace.req_001@cluster-node",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedCtxID string
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtxID = apphttp.GetRequestID(r.Context())
				w.WriteHeader(http.StatusCreated)
			})

			wrapped := apphttp.RequestID(dummyHandler)

			req := httptest.NewRequest(http.MethodPost, "/resource", nil)
			req.Header.Set("X-Request-ID", tc.incomingID)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			outgoingHeader := rec.Header().Get("X-Request-ID")
			if outgoingHeader != tc.incomingID {
				t.Fatalf("expected outgoing X-Request-ID %q, got %q", tc.incomingID, outgoingHeader)
			}

			if capturedCtxID != tc.incomingID {
				t.Fatalf("expected context request ID %q, got %q", tc.incomingID, capturedCtxID)
			}
		})
	}
}

func TestGetRequestID(t *testing.T) {
	t.Run("AC-3: returns empty string when context is nil", func(t *testing.T) {
		got := apphttp.GetRequestID(nil)
		if got != "" {
			t.Fatalf("expected empty string for nil context, got %q", got)
		}
	})

	t.Run("AC-3: returns empty string when context has no request ID", func(t *testing.T) {
		got := apphttp.GetRequestID(context.Background())
		if got != "" {
			t.Fatalf("expected empty string for empty context, got %q", got)
		}
	})

	t.Run("AC-3: returns empty string when context has unrelated keys", func(t *testing.T) {
		type unrelatedKey struct{}
		ctx := context.WithValue(context.Background(), unrelatedKey{}, "unrelated-value")
		got := apphttp.GetRequestID(ctx)
		if got != "" {
			t.Fatalf("expected empty string for context without request ID, got %q", got)
		}
	})
}

func TestLogging_IncludesRequestID(t *testing.T) {
	t.Run("AC-4: Logging middleware includes request_id in structured slog entry when wrapped with RequestID", func(t *testing.T) {
		var logBuf bytes.Buffer
		jsonHandler := slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelInfo})
		logger := slog.New(jsonHandler)
		origLogger := slog.Default()
		slog.SetDefault(logger)
		defer slog.SetDefault(origLogger)

		dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Stack RequestID outside Logging
		handler := apphttp.RequestID(apphttp.Logging(dummyHandler))

		customID := "test-logging-trace-id-789"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.Header.Set("X-Request-ID", customID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var logEntry map[string]any
		if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse log JSON: %v, raw output: %s", err, logBuf.String())
		}

		reqIDVal, ok := logEntry["request_id"]
		if !ok {
			t.Fatalf("expected 'request_id' key in log output, got keys: %v", getMapKeys(logEntry))
		}
		if reqIDVal != customID {
			t.Fatalf("expected request_id %q, got %v", customID, reqIDVal)
		}

		if logEntry["msg"] != "http request" {
			t.Fatalf("expected msg 'http request', got %v", logEntry["msg"])
		}
	})

	t.Run("AC-4: Logging middleware handles context without request_id gracefully", func(t *testing.T) {
		var logBuf bytes.Buffer
		jsonHandler := slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelInfo})
		logger := slog.New(jsonHandler)
		origLogger := slog.Default()
		slog.SetDefault(logger)
		defer slog.SetDefault(origLogger)

		dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Logging without RequestID middleware
		handler := apphttp.Logging(dummyHandler)

		req := httptest.NewRequest(http.MethodGet, "/no-req-id", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var logEntry map[string]any
		if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse log JSON: %v, raw output: %s", err, logBuf.String())
		}

		if logEntry["msg"] != "http request" {
			t.Fatalf("expected msg 'http request', got %v", logEntry["msg"])
		}
	})
}

func TestRecovery_IncludesRequestID(t *testing.T) {
	t.Run("AC-5: Recovery middleware includes request_id in panic log and returns 500 with X-Request-ID", func(t *testing.T) {
		var logBuf bytes.Buffer
		jsonHandler := slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(jsonHandler)
		origLogger := slog.Default()
		slog.SetDefault(logger)
		defer slog.SetDefault(origLogger)

		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("deliberate test panic for recovery verification")
		})

		// Stack RequestID outside Recovery
		handler := apphttp.RequestID(apphttp.Recovery(panicHandler))

		customID := "panic-recovery-trace-456"
		req := httptest.NewRequest(http.MethodGet, "/panic-route", nil)
		req.Header.Set("X-Request-ID", customID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500 InternalServerError, got %d", rec.Code)
		}

		outgoingID := rec.Header().Get("X-Request-ID")
		if outgoingID != customID {
			t.Fatalf("expected outgoing X-Request-ID %q, got %q", customID, outgoingID)
		}

		var logEntry map[string]any
		if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse panic log JSON: %v, raw output: %s", err, logBuf.String())
		}

		reqIDVal, ok := logEntry["request_id"]
		if !ok {
			t.Fatalf("expected 'request_id' key in panic log output, got keys: %v", getMapKeys(logEntry))
		}
		if reqIDVal != customID {
			t.Fatalf("expected request_id %q in panic log, got %v", customID, reqIDVal)
		}

		if logEntry["msg"] != "recovered from panic in http handler" {
			t.Fatalf("expected msg 'recovered from panic in http handler', got %v", logEntry["msg"])
		}
	})
}

func ptr(s string) *string {
	return &s
}

func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Split(strings.Join(keys, ", "), ", ")
}
