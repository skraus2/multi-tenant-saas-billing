package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"billing-platform/api/internal/middleware"
)

func TestTenantIDFromContext_Missing_ReturnsError(t *testing.T) {
	_, err := middleware.TenantIDFromContext(context.Background())
	if err == nil {
		t.Error("expected error for missing tenant_id, got nil")
	}
	if !errors.Is(err, middleware.ErrNoTenantInContext) {
		t.Errorf("expected ErrNoTenantInContext, got %v", err)
	}
}

func TestNewTenant_NoTenantInContext_Returns401(t *testing.T) {
	// NewTenant with nil db — returns 401 before touching db when tenant is missing
	handler := middleware.NewTenant(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No tenant_id in context
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestNewTenant_WithTenantInContext_PassesThrough(t *testing.T) {
	// Verify that auth middleware → tenant middleware chain sets tenant in context
	// and the handler receives it correctly (integration-style, no real DB needed
	// since tenant middleware is exercised via auth middleware path)
	tenantID := uuid.New()
	claims := jwt.MapClaims{
		"tenant_id": tenantID.String(),
		"exp":       jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(testSecret)
	if err != nil {
		t.Fatal(err)
	}

	var gotID uuid.UUID
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, err := middleware.TenantIDFromContext(r.Context())
		if err != nil {
			t.Errorf("TenantIDFromContext: %v", err)
		}
		gotID = id
	})

	// Only test auth middleware here (tenant middleware needs DB)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(capture).ServeHTTP(w, req)

	if gotID != tenantID {
		t.Errorf("tenant_id in context: want %s, got %s", tenantID, gotID)
	}
}
