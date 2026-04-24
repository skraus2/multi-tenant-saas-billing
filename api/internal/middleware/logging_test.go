package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"billing-platform/api/internal/middleware"
)

// capturingHandler is a slog.Handler that records the last logged record.
type capturingHandler struct {
	last  *slog.Record
	attrs []slog.Attr
}

func (h *capturingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *capturingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &capturingHandler{attrs: append(h.attrs, attrs...)}
}
func (h *capturingHandler) WithGroup(_ string) slog.Handler { return h }
func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.last = &r
	return nil
}

func attrValue(r *slog.Record, key string) any {
	var val any
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			val = a.Value.Any()
			return false
		}
		return true
	})
	return val
}

func TestRequestLogger_LogsMethodPathStatus(t *testing.T) {
	handler := &capturingHandler{}
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(slog.Default()) })

	// Wrap a simple 201 handler with RequestID + our logger
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	chain := chimiddleware.RequestID(middleware.NewRequestLogger()(inner))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	if handler.last == nil {
		t.Fatal("expected a log record, got none")
	}

	r := handler.last
	if got := attrValue(r, "method"); got != "GET" {
		t.Errorf("method: want GET, got %v", got)
	}
	if got := attrValue(r, "path"); got != "/health" {
		t.Errorf("path: want /health, got %v", got)
	}
	if got, ok := attrValue(r, "status").(int64); !ok || got != http.StatusCreated {
		t.Errorf("status: want 201, got %v", attrValue(r, "status"))
	}
	if attrValue(r, "latency_ms") == nil {
		t.Error("latency_ms missing")
	}
}

func TestRequestLogger_DefaultStatus200(t *testing.T) {
	handler := &capturingHandler{}
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(slog.Default()) })

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no explicit WriteHeader → defaults to 200
		w.Write([]byte("ok")) //nolint:errcheck
	})
	middleware.NewRequestLogger()(inner).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	if got, ok := attrValue(handler.last, "status").(int64); !ok || got != http.StatusOK {
		t.Errorf("status: want 200, got %v", attrValue(handler.last, "status"))
	}
}
