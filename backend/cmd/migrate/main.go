package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := "status"
	if len(args) > 0 {
		command = args[0]
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", cfg.Database.URL)
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
	dir := goosemigrate.Dir()
	switch command {
	case "baseline-existing":
		versions, err := goosemigrate.BaselineExisting(context.Background(), db)
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			fmt.Println("No existing migrations needed baselining.")
			return nil
		}
		fmt.Printf("Marked existing migrations as applied: %v\n", versions)
		return nil
	case "up":
		if _, err := goosemigrate.BaselineExisting(context.Background(), db); err != nil {
			return err
		}
		return goose.Up(db, dir)
	case "down":
		return goose.Down(db, dir)
	case "status":
		return goose.Status(db, dir)
	case "reset":
		return goose.Reset(db, dir)
	case "version":
		return goose.Version(db, dir)
	default:
		return fmt.Errorf("unsupported command %q; use up, down, status, reset, version, or baseline-existing", command)
	}
}
