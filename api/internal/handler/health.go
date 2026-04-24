package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// HealthResponse is the response body for the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// NewHealth returns an http.HandlerFunc for GET /health.
// It pings the database and returns "ok" (200) or "degraded" (503).
func NewHealth(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		code := http.StatusOK

		if err := db.PingContext(ctx); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(HealthResponse{Status: status}) //nolint:errcheck
	}
}
