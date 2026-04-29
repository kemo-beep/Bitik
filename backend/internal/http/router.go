package httpapi

import (
	"net/http"
	"time"

	"github.com/bitik/backend/internal/adminsvc"
	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/catalogsvc"
	"github.com/bitik/backend/internal/chatsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/mediasvc"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/notificationsvc"
	"github.com/bitik/backend/internal/ordersvc"
	"github.com/bitik/backend/internal/paymentsvc"
	"github.com/bitik/backend/internal/promotionsvc"
	"github.com/bitik/backend/internal/reviewsvc"
	"github.com/bitik/backend/internal/searchsvc"
	"github.com/bitik/backend/internal/sellersvc"
	"github.com/bitik/backend/internal/shippingsvc"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

type VersionInfo struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

type RouterOptions struct {
	RedisClient         *redis.Client
	OpenAPIYAML         []byte
	ExposeMetrics       bool
	TracingEnabled      bool
	AuthService         *authsvc.Service
	AdminService        *adminsvc.Service
	CatalogService      *catalogsvc.Service
	MediaService        *mediasvc.Service
	OrderService        *ordersvc.Service
	SellerService       *sellersvc.Service
	PaymentService      *paymentsvc.Service
	ShippingService     *shippingsvc.Service
	SearchService       *searchsvc.Service
	PromotionsService   *promotionsvc.Service
	ReviewService       *reviewsvc.Service
	NotificationService *notificationsvc.Service
	ChatService         *chatsvc.Service
}

func NewRouter(cfg config.Config, logger *zap.Logger, readiness *Readiness, opts RouterOptions) *gin.Engine {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.RequestID())
	if opts.TracingEnabled {
		router.Use(otelgin.Middleware(cfg.App.Name))
	}
	router.Use(
		middleware.Recovery(logger),
		middleware.Logger(logger),
		middleware.BodyLimit(cfg.Security.MaxRequestBodyBytes),
		middleware.CORS(cfg.CORS),
		middleware.RateLimit(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst),
		middleware.AuthParser(),
		middleware.Idempotency(opts.RedisClient, cfg.Idempotency),
	)

	router.NoRoute(func(c *gin.Context) {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Route not found.")
	})

	router.GET("/health", func(c *gin.Context) {
		apiresponse.OK(c, gin.H{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		checks, ready := readiness.Check(c.Request.Context())
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		apiresponse.Respond(c, status, gin.H{"checks": checks}, nil)
	})

	if opts.ExposeMetrics {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	router.GET("/version", func(c *gin.Context) {
		apiresponse.OK(c, VersionInfo{
			Name:        cfg.App.Name,
			Environment: cfg.App.Environment,
			Version:     cfg.App.Version,
			Commit:      cfg.App.Commit,
		})
	})

	if len(opts.OpenAPIYAML) > 0 {
		spec := opts.OpenAPIYAML
		router.GET("/openapi.yaml", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/yaml", spec)
		})
	}

	router.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerDocsHTML))
	})

	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/docs")
	})

	v1 := router.Group("/api/v1")
	v1.GET("", func(c *gin.Context) {
		apiresponse.OK(c, gin.H{"name": cfg.App.Name, "version": cfg.App.Version})
	})

	if opts.AuthService != nil {
		opts.AuthService.RegisterRoutes(v1)
	}
	if opts.AdminService != nil {
		opts.AdminService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.CatalogService != nil {
		opts.CatalogService.RegisterRoutes(v1)
	}
	if opts.MediaService != nil {
		opts.MediaService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.SellerService != nil {
		opts.SellerService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.OrderService != nil {
		opts.OrderService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.PaymentService != nil {
		opts.PaymentService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.ShippingService != nil {
		opts.ShippingService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.SearchService != nil {
		opts.SearchService.RegisterRoutes(v1)
	}
	if opts.PromotionsService != nil {
		opts.PromotionsService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.ReviewService != nil {
		opts.ReviewService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.NotificationService != nil {
		opts.NotificationService.RegisterRoutes(v1, opts.AuthService)
	}
	if opts.ChatService != nil {
		opts.ChatService.RegisterRoutes(v1, opts.AuthService)
	}

	return router
}
