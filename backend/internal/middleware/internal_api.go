package middleware

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/config"
	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

func RequireInternalAPI(cfg config.Config) gin.HandlerFunc {
	allowed := parseCIDRs(cfg.InternalAPI.AllowedCIDRs)
	want := strings.TrimSpace(cfg.InternalAPI.Token)

	return func(c *gin.Context) {
		got := strings.TrimSpace(c.GetHeader(internalTokenHeader))
		if want == "" || got == "" || got != want {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Internal token is missing or invalid.")
			c.Abort()
			return
		}

		if len(allowed) > 0 {
			ip, err := netip.ParseAddr(strings.TrimSpace(c.ClientIP()))
			if err != nil {
				apiresponse.Error(c, http.StatusForbidden, "forbidden", "Client IP is not allowed.")
				c.Abort()
				return
			}
			ok := false
			for _, p := range allowed {
				if p.Contains(ip) {
					ok = true
					break
				}
			}
			if !ok {
				apiresponse.Error(c, http.StatusForbidden, "forbidden", "Client IP is not allowed.")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func parseCIDRs(items []string) []netip.Prefix {
	out := make([]netip.Prefix, 0, len(items))
	for _, raw := range items {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		p, err := netip.ParsePrefix(raw)
		if err != nil {
			// accept bare IPs as /32 or /128
			if ip, err2 := netip.ParseAddr(raw); err2 == nil {
				bits := 32
				if ip.Is6() {
					bits = 128
				}
				out = append(out, netip.PrefixFrom(ip, bits))
			}
			continue
		}
		out = append(out, p)
	}
	return out
}
