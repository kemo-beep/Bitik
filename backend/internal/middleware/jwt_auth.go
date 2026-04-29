package middleware

import (
	"net/http"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/jwtutil"
	"github.com/gin-gonic/gin"
)

// RequireBearerJWT validates Authorization: Bearer and sets AuthUserIDKey (uuid string) and AuthRolesKey.
func RequireBearerJWT(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.GetHeader("Authorization"))
		token, ok := strings.CutPrefix(raw, "Bearer ")
		if !ok || token == "" {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "A valid bearer access token is required.")
			c.Abort()
			return
		}
		claims, err := jwtutil.Parse(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, token)
		if err != nil {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Invalid or expired access token.")
			c.Abort()
			return
		}
		sub, err := jwtutil.SubjectUUID(claims)
		if err != nil {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Invalid access token subject.")
			c.Abort()
			return
		}
		c.Set(AuthUserIDKey, sub.String())
		c.Set(AuthRolesKey, claims.Roles)
		c.Next()
	}
}
