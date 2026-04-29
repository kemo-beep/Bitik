package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestWeightByRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		if w := requestWeight(c); w != 3 {
			t.Fatalf("login weight=%d want=3", w)
		}
		c.Status(http.StatusOK)
	})
	r.PATCH("/api/v1/buyer/orders/:order_id", func(c *gin.Context) {
		if w := requestWeight(c); w != 2 {
			t.Fatalf("mutation weight=%d want=2", w)
		}
		c.Status(http.StatusOK)
	})
	r.GET("/api/v1/catalog/products", func(c *gin.Context) {
		if w := requestWeight(c); w != 1 {
			t.Fatalf("read weight=%d want=1", w)
		}
		c.Status(http.StatusOK)
	})

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/auth/login"},
		{http.MethodPatch, "/api/v1/buyer/orders/1"},
		{http.MethodGet, "/api/v1/catalog/products"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s expected 200 got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestRateLimitRouteWeightedBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthParser(), RateLimit(1, 2))
	r.POST("/api/v1/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/catalog/products", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Risk endpoint consumes weight 3 and should be blocked immediately.
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req1.RemoteAddr = "10.10.10.10:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusTooManyRequests {
		t.Fatalf("expected login route to be limited, got %d", w1.Code)
	}

	// Read endpoint weight 1 should pass first request.
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products", nil)
	req2.RemoteAddr = "11.11.11.11:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected read route 200, got %d", w2.Code)
	}
}
