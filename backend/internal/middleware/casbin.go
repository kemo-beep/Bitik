package middleware

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/rbac"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// RequireCasbinHTTP allows the request when any JWT role matches a Casbin policy for path + method.
func RequireCasbinHTTP(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			apiresponse.Error(c, http.StatusServiceUnavailable, "rbac_unavailable", "Authorization is not configured.")
			c.Abort()
			return
		}
		raw, _ := c.Get(AuthRolesKey)
		roles, _ := raw.([]string)
		if len(roles) == 0 {
			apiresponse.Error(c, http.StatusForbidden, "forbidden", "No roles on token.")
			c.Abort()
			return
		}
		if !rbac.AllowAnyRole(enforcer, roles, c.Request.URL.Path, c.Request.Method) {
			apiresponse.Error(c, http.StatusForbidden, "forbidden", "Insufficient permissions for this resource.")
			c.Abort()
			return
		}
		c.Next()
	}
}
