package middleware

import (
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("http_request",
			zap.String("request_id", requestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

func requestID(c *gin.Context) string {
	if value, ok := c.Get(apiresponse.RequestIDContextKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return c.GetHeader("X-Request-ID")
}
