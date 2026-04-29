package middleware

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
)

func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 || c.Request == nil || c.Request.Body == nil {
			c.Next()
			return
		}
		if c.Request.ContentLength > maxBytes {
			apiresponse.Error(c, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body exceeds allowed limit.")
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
