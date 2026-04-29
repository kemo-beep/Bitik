package middleware

import (
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
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: allowCreds,
		MaxAge:           12 * time.Hour,
	})
}
