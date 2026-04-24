// migrate runs all SQL migration files from the migrations/ directory in
// alphabetical order, each in its own transaction. Already-applied migrations
// are tracked in schema_migrations and skipped on re-runs. After all migrations
// succeed, it sets the billing_app role password from POSTGRES_APP_PASSWORD.
//
// Uses DATABASE_URL (superuser billing) — NOT DATABASE_APP_URL.
// Run: go run ./cmd/migrate
package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations completed successfully")
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	// Find migration files
	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		slog.Info("no migration files found")
		return nil
	}

	// Run each migration in its own transaction (skip already-applied ones)
	for _, file := range files {
		filename := filepath.Base(file)

		applied, err := isMigrationApplied(db, filename)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", filename, err)
		}
		if applied {
			slog.Info("migration already applied, skipping", "file", filename)
			continue
		}

		if err := runMigration(db, file, filename); err != nil {
			return fmt.Errorf("migration %s: %w", filename, err)
		}
	}

	// Set billing_app password after all migrations succeed
	if appPassword := os.Getenv("POSTGRES_APP_PASSWORD"); appPassword != "" {
		// ALTER ROLE does not support $1 placeholders — safe here since value
		// comes from env var, not user input.
		if _, err := db.Exec(fmt.Sprintf("ALTER ROLE billing_app PASSWORD '%s'", appPassword)); err != nil {
			return fmt.Errorf("set billing_app password: %w", err)
		}
		slog.Info("billing_app password set")
	} else {
		slog.Warn("POSTGRES_APP_PASSWORD not set — billing_app has no password")
	}

	return nil
}

// ensureMigrationsTable creates the schema_migrations tracking table if it
// does not already exist. Uses the superuser connection, so it runs before RLS.
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename   TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	return err
}

func isMigrationApplied(db *sql.DB, filename string) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE filename = $1",
		filename,
	).Scan(&count)
	return count > 0, err
}

func runMigration(db *sql.DB, file, filename string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback() //nolint:errcheck
		return fmt.Errorf("exec: %w", err)
	}

	if _, err := tx.Exec(
		"INSERT INTO schema_migrations(filename) VALUES ($1)",
		filename,
	); err != nil {
		tx.Rollback() //nolint:errcheck
		return fmt.Errorf("mark applied: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	slog.Info("migration applied", "file", filename)
	return nil
}
