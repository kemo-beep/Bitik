package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimit(requestsPerSecond float64, burst int) gin.HandlerFunc {
	if requestsPerSecond <= 0 || burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	limiters := map[string]*rate.Limiter{}
	var mu sync.Mutex

	return func(c *gin.Context) {
		key := limitKey(c)
		weight := requestWeight(c)
		mu.Lock()
		limiter, ok := limiters[key]
		if !ok {
			limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
			limiters[key] = limiter
		}
		mu.Unlock()

		if !limiter.AllowN(time.Now(), weight) {
			apiresponse.Error(c, http.StatusTooManyRequests, "rate_limited", "Too many requests.")
			c.Abort()
			return
		}

		c.Next()
	}
}

func limitKey(c *gin.Context) string {
	userID := "anon"
	if v, ok := c.Get(UserIDKey); ok {
		if s, ok := v.(string); ok && s != "" {
			userID = s
		}
	}
	return fmt.Sprintf("ip:%s|user:%s|route:%s|method:%s", c.ClientIP(), userID, c.FullPath(), c.Request.Method)
}

func requestWeight(c *gin.Context) int {
	path := c.FullPath()
	switch {
	case path == "" || path == "/":
		return 1
	case path == "/api/v1/auth/login" || path == "/api/v1/auth/refresh-token" || path == "/api/v1/auth/send-phone-otp":
		return 3
	case c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete:
		return 2
	default:
		return 1
	}
}
