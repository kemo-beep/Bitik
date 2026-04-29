package goosemigrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

// Dir resolves the Goose migrations directory for both `go run` from `backend/`
// and from the monorepo root. Override with BITIK_MIGRATIONS_DIR.
func Dir() string {
	if d := os.Getenv("BITIK_MIGRATIONS_DIR"); d != "" {
		return d
	}
	for _, c := range []string{"db/migrations", "backend/db/migrations"} {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return "db/migrations"
}

type existingMigrationSentinel struct {
	Version int64
	Table   string
}

var existingMigrationSentinels = []existingMigrationSentinel{
	{Version: 20260428042113, Table: "users"},
	{Version: 20260428042114, Table: "roles"},
	{Version: 20260428042115, Table: "sellers"},
	{Version: 20260428042117, Table: "products"},
	{Version: 20260428042118, Table: "inventory_items"},
	{Version: 20260428042119, Table: "orders"},
	{Version: 20260428042120, Table: "payments"},
	{Version: 20260428042121, Table: "shipments"},
	{Version: 20260428042122, Table: "vouchers"},
	{Version: 20260428042123, Table: "notifications"},
	{Version: 20260428042124, Table: "platform_settings"},
	{Version: 20260428042125, Table: "audit_logs"},
	{Version: 20260428043610, Table: "user_sessions"},
}

// BaselineExisting records migrations whose sentinel tables already exist but
// whose Goose version rows are missing. This repairs development/staging DBs
// created before Goose metadata was aligned, without dropping user data.
func BaselineExisting(ctx context.Context, db *sql.DB) ([]int64, error) {
	applied := make([]int64, 0)
	for _, sentinel := range existingMigrationSentinels {
		exists, err := tableExists(ctx, db, sentinel.Table)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		inserted, err := markVersionApplied(ctx, db, sentinel.Version)
		if err != nil {
			return nil, err
		}
		if inserted {
			applied = append(applied, sentinel.Version)
		}
	}
	return applied, nil
}

func tableExists(ctx context.Context, db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = $1
		)
	`, tableName).Scan(&exists)
	return exists, err
}

func markVersionApplied(ctx context.Context, db *sql.DB, version int64) (bool, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO goose_db_version (version_id, is_applied)
		SELECT $1, TRUE
		WHERE NOT EXISTS (
			SELECT 1
			FROM goose_db_version
			WHERE version_id = $1
			  AND is_applied = TRUE
		)
	`, version)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// RunFromDSN opens a connection, applies pending migrations with Goose, then closes.
func RunFromDSN(databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("database URL is empty")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return err
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if _, err := BaselineExisting(context.Background(), db); err != nil {
		return err
	}
	return goose.Up(db, Dir())
}
