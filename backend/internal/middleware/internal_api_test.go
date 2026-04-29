package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitik/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func TestRequireInternalAPI_TokenRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{InternalAPI: config.InternalAPIConfig{Token: "secret"}}
	r.POST("/internal", RequireInternalAPI(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireInternalAPI_AcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{InternalAPI: config.InternalAPIConfig{Token: "secret"}}
	r.POST("/internal", RequireInternalAPI(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Token", "secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireInternalAPI_CIDRAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		InternalAPI: config.InternalAPIConfig{
			Token:        "secret",
			AllowedCIDRs: []string{"10.0.0.0/8"},
		},
	}
	r.POST("/internal", RequireInternalAPI(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/internal", nil)
	req.Header.Set("X-Internal-Token", "secret")
	req.RemoteAddr = "192.168.1.10:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
