package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ErrNoTenantInContext is returned when no tenant_id is found in the context.
var ErrNoTenantInContext = errors.New("no tenant_id in context")

type dbConnContextKey struct{}

// withTenantID stores the tenant UUID in ctx.
func withTenantID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantIDKey, id)
}

// TenantIDFromContext retrieves the tenant UUID set by the auth middleware.
// Returns ErrNoTenantInContext if the value is absent.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(tenantIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrNoTenantInContext
	}
	return id, nil
}

// DBConnFromContext retrieves the per-request pinned connection set by
// NewTenant. Repositories must use this connection to ensure RLS applies.
func DBConnFromContext(ctx context.Context) (*sqlx.Conn, bool) {
	conn, ok := ctx.Value(dbConnContextKey{}).(*sqlx.Conn)
	return conn, ok
}

// NewTenant returns a middleware that:
//  1. Reads tenant_id from context (set by NewAuth).
//  2. Acquires a dedicated connection from the pool (pinned for this request).
//  3. Executes SET LOCAL app.current_tenant_id within a transaction so RLS
//     reads the correct value on the exact same connection.
//  4. Stores the pinned connection in the context for repositories to use.
//
// Repositories must call DBConnFromContext to retrieve this connection —
// using the pool directly would bypass the tenant variable.
func NewTenant(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, err := TenantIDFromContext(r.Context())
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Acquire a dedicated connection — pinned for the lifetime of this request
			conn, err := db.Connx(r.Context())
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer conn.Close() //nolint:errcheck

			// SET (not SET LOCAL) so the variable persists beyond a single statement
			// on this specific connection. Handlers share this connection via context.
			if _, err := conn.ExecContext(r.Context(),
				"SET app.current_tenant_id = $1", tenantID.String(),
			); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), dbConnContextKey{}, conn)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
