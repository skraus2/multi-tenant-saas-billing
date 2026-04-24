package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"billing-platform/api/internal/middleware"
)

var testSecret = []byte("test-secret-32-bytes-long-enough!")

func makeToken(t *testing.T, tenantID string, secret []byte, expiry time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"tenant_id": tenantID,
		"exp":       jwt.NewNumericDate(expiry),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return str
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth_ValidJWT_Passes(t *testing.T) {
	tenantID := uuid.New().String()
	token := makeToken(t, tenantID, testSecret, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuth_NoHeader_Returns401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	tenantID := uuid.New().String()
	token := makeToken(t, tenantID, testSecret, time.Now().Add(-time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_WrongSecret_Returns401(t *testing.T) {
	tenantID := uuid.New().String()
	token := makeToken(t, tenantID, []byte("other-secret"), time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidTenantIDClaim_Returns401(t *testing.T) {
	// tenant_id is not a valid UUID
	token := makeToken(t, "not-a-uuid", testSecret, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_AlgNone_Returns401(t *testing.T) {
	// Construct a token manually with alg:none to test the algorithm check
	tenantID := uuid.New().String()
	claims := jwt.MapClaims{
		"tenant_id": tenantID,
		"exp":       jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	middleware.NewAuth(testSecret)(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for alg:none, got %d", w.Code)
	}
}

func TestAuth_TenantIDInContext(t *testing.T) {
	tenantID := uuid.New()
	token := makeToken(t, tenantID.String(), testSecret, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	var gotID uuid.UUID
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, err := middleware.TenantIDFromContext(r.Context())
		if err != nil {
			t.Errorf("TenantIDFromContext: %v", err)
		}
		gotID = id
	})

	middleware.NewAuth(testSecret)(capture).ServeHTTP(w, req)

	if gotID != tenantID {
		t.Errorf("context tenant_id: want %s, got %s", tenantID, gotID)
	}
}
