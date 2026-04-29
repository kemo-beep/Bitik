//go:build integration

package authsvc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/db"
	authstore "github.com/bitik/backend/internal/store/auth"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Run with: go test -tags=integration ./internal/authsvc -count=1
// Requires BITIK_DATABASE_URL and a migrated database. Optional BITIK_REDIS_ADDR.

func TestRefreshRotationPostgres(t *testing.T) {
	dsn := os.Getenv("BITIK_DATABASE_URL")
	if dsn == "" {
		t.Skip("BITIK_DATABASE_URL not set")
	}
	cfg := config.Config{}
	cfg.Auth.JWTSecret = "integration-test-secret-32bytes!!"
	cfg.Auth.JWTIssuer = "bitik-test"
	cfg.Auth.AccessTokenTTL = time.Minute
	cfg.Auth.RefreshTokenTTL = 24 * time.Hour
	cfg.Auth.PublicBaseURL = "http://127.0.0.1:8080"

	ctx := context.Background()
	pool, err := db.Connect(ctx, config.DatabaseConfig{URL: dsn, MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var r *redis.Client
	if addr := os.Getenv("BITIK_REDIS_ADDR"); addr != "" {
		r = redis.NewClient(&redis.Options{Addr: addr, DB: 0})
		defer r.Close()
		if err := r.Ping(ctx).Err(); err != nil {
			t.Fatal(err)
		}
	}

	svc := NewService(cfg, zap.NewNop(), pool, r, nil)

	meta := clientMeta{UserAgent: "integration-test"}
	email := "integration_refresh_" + time.Now().Format("150405") + "@bitik.test"
	pair1, uid, err := svc.Register(ctx, email, "password123", nil, meta)
	if err != nil {
		t.Fatal(err)
	}
	pair2, err := svc.Refresh(ctx, pair1.RefreshToken, meta)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(ctx, pair1.RefreshToken, meta)
	if err == nil {
		t.Fatal("expected error reusing old refresh token after rotation")
	}
	if pair2.AccessToken == "" || pair2.RefreshToken == "" {
		t.Fatal("missing tokens")
	}

	sessions, err := svc.auth.ListUserSessionsForUser(ctx, pgxutil.UUID(uid))
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) == 0 {
		t.Fatal("expected session")
	}
	if err := svc.auth.RevokeUserSession(ctx, authstore.RevokeUserSessionParams{
		ID:     sessions[0].ID,
		UserID: pgxutil.UUID(uid),
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.auth.RevokeRefreshTokensForSession(ctx, authstore.RevokeRefreshTokensForSessionParams{
		SessionID: sessions[0].ID,
		UserID:    pgxutil.UUID(uid),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(ctx, pair2.RefreshToken, meta); err == nil {
		t.Fatal("expected revoked session refresh token to fail")
	}
}
