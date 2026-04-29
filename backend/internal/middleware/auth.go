package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	AuthTokenKey  = "auth.token"
	UserIDKey     = "auth.user_id"
	AuthUserIDKey = "auth.user_uuid"
	AuthRolesKey  = "auth.roles"
)

func AuthParser() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if token, ok := strings.CutPrefix(header, "Bearer "); ok && token != "" {
			c.Set(AuthTokenKey, token)
		}

		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set(UserIDKey, userID)
		}

		c.Next()
	}
}
