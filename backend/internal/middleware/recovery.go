package middleware

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic_recovered",
			zap.Any("panic", recovered),
			zap.String("request_id", requestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
		c.Abort()
	})
}
