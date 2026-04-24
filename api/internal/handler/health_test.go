package handler_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"

	"billing-platform/api/internal/handler"
)

// mockDriver + mockConn implement driver.Driver, driver.Conn, and driver.Pinger
// to create a fake sqlx.DB without real Postgres. pingErr controls PingContext.
type mockDriver struct{ pingErr error }
type mockConn struct{ pingErr error }

func (d *mockDriver) Open(_ string) (driver.Conn, error)  { return &mockConn{pingErr: d.pingErr}, nil }
func (c *mockConn) Prepare(_ string) (driver.Stmt, error) { return nil, nil }
func (c *mockConn) Close() error                          { return nil }
func (c *mockConn) Begin() (driver.Tx, error)             { return nil, nil }

// Ping implements driver.Pinger so db.PingContext uses our error.
func (c *mockConn) Ping(_ context.Context) error { return c.pingErr }

func newMockDB(t *testing.T, pingErr error) *sqlx.DB {
	t.Helper()
	driverName := "mock_" + t.Name()
	sql.Register(driverName, &mockDriver{pingErr: pingErr})
	db, err := sqlx.Open(driverName, "")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestHealth_DBUp_ReturnsOK(t *testing.T) {
	db := newMockDB(t, nil)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.NewHealth(db)(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestHealth_DBDown_ReturnsDegraded(t *testing.T) {
	db := newMockDB(t, errors.New("connection refused"))
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.NewHealth(db)(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "degraded" {
		t.Errorf("expected status=degraded, got %q", body["status"])
	}
}

func TestHealth_ReturnsJSONContentType(t *testing.T) {
	db := newMockDB(t, nil)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.NewHealth(db)(w, req)

	ct := w.Result().Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
