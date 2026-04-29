package db

import (
	"context"
	"fmt"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = cfg.MaxOpenConns
	poolConfig.MinConns = cfg.MaxIdleConns
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	if cfg.QueryTimeout > 0 {
		poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = durationToMillis(cfg.QueryTimeout)
	}

	connectCtx := ctx
	cancel := func() {}
	if cfg.ConnectTimeout > 0 {
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
	}
	defer cancel()
	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func durationToMillis(d time.Duration) string {
	ms := d.Milliseconds()
	if ms <= 0 {
		ms = 1
	}
	return fmt.Sprintf("%d", ms)
}
