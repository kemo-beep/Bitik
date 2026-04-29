//go:build integration

package adminsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestPhase9_ModerationSettingsAndAnalyticsFlow(t *testing.T) {
	dsn := os.Getenv("BITIK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BITIK_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := goosemigrate.RunFromDSN(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	cleanupPhase9(t, ctx, pool)
	adminID := mustInsertPhase9User(t, ctx, pool, "phase9-admin-"+uuid.NewString()+"@example.com")
	targetUserID := mustInsertPhase9User(t, ctx, pool, "phase9-target-"+uuid.NewString()+"@example.com")

	var reportID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO moderation_reports (reporter_user_id, target_type, target_id, reason, status)
		VALUES ($1, 'product', $2, 'spam listing', 'open')
		RETURNING id
	`, targetUserID, uuid.New()).Scan(&reportID); err != nil {
		t.Fatalf("insert moderation report: %v", err)
	}

	svc := NewService(config.Config{}, zap.NewNop(), pool, nil)
	gin.SetMode(gin.TestMode)

	// Create moderation case.
	{
		body, _ := json.Marshal(map[string]any{
			"report_id": reportID.String(),
			"status":    "under_review",
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/admin/moderation/cases", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.AuthUserIDKey, adminID)

		svc.HandleAdminCreateModerationCase(c)
		if w.Code != 200 {
			t.Fatalf("create moderation case expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Upsert a platform setting.
	{
		body, _ := json.Marshal(map[string]any{
			"value":       map[string]any{"enabled": true},
			"description": "phase9 integration",
			"is_public":   false,
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/admin/settings/platform/phase9.setting", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "key", Value: "phase9.setting"}}
		c.Set(middleware.AuthUserIDKey, adminID)

		svc.HandleAdminUpsertPlatformSetting(c)
		if w.Code != 200 {
			t.Fatalf("upsert platform setting expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Upsert a feature flag.
	{
		body, _ := json.Marshal(map[string]any{
			"enabled":     true,
			"description": "phase9 feature flag",
			"rules":       map[string]any{"rollout": 100},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/admin/settings/feature-flags/phase9.flag", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "key", Value: "phase9.flag"}}
		c.Set(middleware.AuthUserIDKey, adminID)

		svc.HandleAdminUpsertFeatureFlag(c)
		if w.Code != 200 {
			t.Fatalf("upsert feature flag expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Ingest event then process queue and verify rollup is written.
	{
		body, _ := json.Marshal(map[string]any{
			"event_name":  "product.view",
			"entity_type": "product",
			"entity_id":   uuid.NewString(),
			"metadata":    map[string]any{"source": "integration"},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/admin/analytics/events", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.AuthUserIDKey, adminID)
		c.Request.RemoteAddr = "127.0.0.1:1234"

		svc.HandleIngestAnalyticsEvent(c)
		if w.Code != 202 {
			t.Fatalf("ingest event expected 202 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Non-admin path: POST /api/v1/analytics/events (any authenticated user; no admin Casbin).
	{
		body, _ := json.Marshal(map[string]any{
			"event_name": "checkout.start",
			"metadata":   map[string]any{"source": "non_admin_analytics_path"},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/analytics/events", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.AuthUserIDKey, targetUserID)
		c.Request.RemoteAddr = "127.0.0.1:2345"

		svc.HandleIngestAnalyticsEvent(c)
		if w.Code != 202 {
			t.Fatalf("non-admin analytics ingest expected 202 got %d body=%s", w.Code, w.Body.String())
		}
	}
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/internal/jobs/process-analytics-events?limit=20", nil)

		svc.HandleProcessAnalyticsEvents(c)
		if w.Code != 200 {
			t.Fatalf("process analytics expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/admin/logs/audit", nil)
		c.Set(middleware.AuthUserIDKey, adminID)

		svc.HandleAdminListAuditLogs(c)
		if w.Code != 200 {
			t.Fatalf("list audit logs expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/admin/logs/activity", nil)
		c.Set(middleware.AuthUserIDKey, adminID)

		svc.HandleAdminListAdminActivityLogs(c)
		if w.Code != 200 {
			t.Fatalf("list admin activity logs expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/admin/dashboard/overview", nil)
		c.Set(middleware.AuthUserIDKey, adminID)

		svc.HandleAdminDashboardOverview(c)
		if w.Code != 200 {
			t.Fatalf("dashboard overview expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	var metricCount int64
	if err := pool.QueryRow(ctx, `SELECT COUNT(1)::bigint FROM admin_metrics_daily WHERE event_name='product.view'`).Scan(&metricCount); err != nil {
		t.Fatalf("count metrics: %v", err)
	}
	if metricCount == 0 {
		t.Fatalf("expected admin_metrics_daily rows for product.view")
	}
}

func mustInsertPhase9User(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, status, email_verified) VALUES ($1, 'active', true) RETURNING id`, email).Scan(&id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func cleanupPhase9(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	stmts := []string{
		`DELETE FROM admin_metrics_daily`,
		`DELETE FROM analytics_event_queue`,
		`DELETE FROM event_logs`,
		`DELETE FROM admin_activity_logs`,
		`DELETE FROM audit_logs`,
		`DELETE FROM moderation_cases`,
		`DELETE FROM moderation_reports`,
		`DELETE FROM platform_settings`,
		`DELETE FROM feature_flags`,
		`DELETE FROM users`,
	}
	for _, stmt := range stmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("cleanup %q: %v", stmt, err)
		}
	}
}
