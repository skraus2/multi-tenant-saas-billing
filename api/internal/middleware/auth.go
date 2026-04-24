package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const tenantIDKey contextKey = "tenant_id"

// BillingClaims are the JWT claims we expect in every token.
type BillingClaims struct {
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// NewAuth returns a middleware that validates a Bearer JWT and puts the
// tenant_id into the request context. Returns 401 on any failure.
func NewAuth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearer(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := &BillingClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				// Prevent alg:none and RSA confusion attacks
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tenantID, err := uuid.Parse(claims.TenantID)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = withTenantID(ctx, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearer extracts the token string from "Authorization: Bearer <token>".
func extractBearer(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid Authorization header format")
	}
	return strings.TrimSpace(parts[1]), nil
}
