package middleware

import (
	"net/url"
	"strings"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowCreds := true
	for _, origin := range cfg.AllowedOrigins {
		if strings.TrimSpace(origin) == "*" {
			allowCreds = false
			break
		}
	}
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		AllowOriginFunc:  allowLocalOrigin(cfg.AllowedOrigins),
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: allowCreds,
		MaxAge:           12 * time.Hour,
	})
}

func allowLocalOrigin(allowed []string) func(string) bool {
	return func(origin string) bool {
		for _, item := range allowed {
			if strings.TrimSpace(item) == origin {
				return true
			}
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		host := u.Hostname()
		return host == "localhost" || host == "127.0.0.1" || host == "::1"
	}
}
