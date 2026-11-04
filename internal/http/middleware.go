package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const HeaderXRequestID = "X-Request-ID"

type contextKey string

const requestIDKey contextKey = "request_id"

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get(HeaderXRequestID))
		if reqID == "" {
			reqID = uuid.NewString()
		}

		w.Header().Set(HeaderXRequestID, reqID)
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrappedWriter.statusCode),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("remote_addr", r.RemoteAddr),
		}
		if reqID := GetRequestID(r.Context()); reqID != "" {
			attrs = append(attrs, slog.String("request_id", reqID))
		}

		slog.LogAttrs(r.Context(), slog.LevelInfo, "http request", attrs...)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				attrs := []slog.Attr{
					slog.Any("error", rec),
					slog.String("path", r.URL.Path),
				}
				if reqID := GetRequestID(r.Context()); reqID != "" {
					attrs = append(attrs, slog.String("request_id", reqID))
				}

				slog.LogAttrs(r.Context(), slog.LevelError, "recovered from panic in http handler", attrs...)

				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

