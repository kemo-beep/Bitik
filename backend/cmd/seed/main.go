package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/platform/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	files, err := filepath.Glob("db/seeds/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		fmt.Printf("seeded %s\n", file)
	}

	return nil
}
