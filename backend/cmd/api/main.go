package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitik/backend/internal/adminsvc"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/catalogsvc"
	"github.com/bitik/backend/internal/chatsvc"
	"github.com/bitik/backend/internal/config"
	httpapi "github.com/bitik/backend/internal/http"
	"github.com/bitik/backend/internal/mediasvc"
	"github.com/bitik/backend/internal/notificationsvc"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/openapi"
	"github.com/bitik/backend/internal/ordersvc"
	"github.com/bitik/backend/internal/paymentsvc"
	"github.com/bitik/backend/internal/platform/cache"
	"github.com/bitik/backend/internal/platform/db"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	"github.com/bitik/backend/internal/platform/observability"
	"github.com/bitik/backend/internal/platform/queue"
	platformsearch "github.com/bitik/backend/internal/platform/search"
	platformstorage "github.com/bitik/backend/internal/platform/storage"
	"github.com/bitik/backend/internal/promotionsvc"
	"github.com/bitik/backend/internal/rbac"
	"github.com/bitik/backend/internal/reviewsvc"
	"github.com/bitik/backend/internal/searchsvc"
	"github.com/bitik/backend/internal/sellersvc"
	"github.com/bitik/backend/internal/shippingsvc"
	rbacstore "github.com/bitik/backend/internal/store/rbac"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	version = "dev"
	commit  = "local"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	cfg.App.Version = valueOrDefault(cfg.App.Version, version)
	cfg.App.Commit = valueOrDefault(cfg.App.Commit, commit)

	logger, err := observability.NewLogger(cfg.Observability)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	rootCtx := context.Background()
	shutdownTracer, err := observability.InitTracing(
		rootCtx,
		cfg.Observability,
		cfg.App.Name,
		cfg.App.Version,
		cfg.App.Environment,
	)
	if err != nil {
		logger.Fatal("tracing_init_failed", zap.Error(err))
	}

	readiness := httpapi.NewReadiness()

	var pgPool *pgxpool.Pool
	if pool, err := db.Connect(rootCtx, cfg.Database); err != nil {
		logger.Warn("database_unavailable", zap.Error(err))
		readiness.Register("postgres", httpapi.Unavailable("postgres"))
	} else {
		pgPool = pool
		defer pgPool.Close()
		readiness.Register("postgres", pgPool.Ping)
		if cfg.Database.AutoMigrate {
			if err := goosemigrate.RunFromDSN(cfg.Database.URL); err != nil {
				logger.Fatal("database_migrate_failed", zap.Error(err))
			}
			logger.Info("database_migrations_applied", zap.String("dir", goosemigrate.Dir()))
		}
	}

	var redisClient *redis.Client
	if client, err := cache.Connect(rootCtx, cfg.Redis); err != nil {
		logger.Warn("redis_unavailable", zap.Error(err))
		readiness.Register("redis", httpapi.Unavailable("redis"))
	} else {
		redisClient = client
		defer redisClient.Close()
		readiness.Register("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		})
	}

	readiness.Register("opensearch", func(ctx context.Context) error {
		return platformsearch.Check(ctx, cfg.Search)
	})

	var minioClient *minio.Client
	if client, err := platformstorage.Connect(rootCtx, cfg.Storage); err != nil {
		logger.Warn("storage_unavailable", zap.Error(err))
		readiness.Register("storage", httpapi.Unavailable("storage"))
	} else {
		minioClient = client
		readiness.Register("storage", func(ctx context.Context) error {
			_, err := minioClient.BucketExists(ctx, cfg.Storage.Bucket)
			return err
		})
	}

	var authSvc *authsvc.Service
	var adminSvc *adminsvc.Service
	var catalogSvc *catalogsvc.Service
	var mediaSvc *mediasvc.Service
	var orderSvc *ordersvc.Service
	var paymentSvc *paymentsvc.Service
	var shippingSvc *shippingsvc.Service
	var sellerSvc *sellersvc.Service
	var searchSvc *searchsvc.Service
	var promoSvc *promotionsvc.Service
	var reviewSvc *reviewsvc.Service
	var notifySvc *notificationsvc.Service
	var chatSvc *chatsvc.Service
	var queueProducer *queue.Producer
	if qBroker, err := queue.NewBroker(cfg.RabbitMQ); err != nil {
		logger.Warn("queue_unavailable", zap.Error(err))
	} else {
		if err := qBroker.SetupTopology(); err != nil {
			logger.Warn("queue_topology_failed", zap.Error(err))
		} else {
			queueProducer = queue.NewProducer(qBroker)
		}
		defer qBroker.Close()
	}
	if pgPool != nil {
		enf, err := rbac.NewEnforcer(rootCtx, rbacstore.New(pgPool))
		if err != nil {
			logger.Fatal("casbin_init_failed", zap.Error(err))
		}
		authSvc = authsvc.NewService(cfg, logger, pgPool, redisClient, enf)
		adminSvc = adminsvc.NewService(cfg, logger, pgPool, readiness.Check)
		catalogSvc = catalogsvc.NewService(cfg, logger, pgPool)
		mediaOpts := []mediasvc.Option{}
		if queueProducer != nil {
			mediaOpts = append(mediaOpts, mediasvc.WithImageProcessor(queueProducer))
		}
		mediaSvc = mediasvc.NewService(cfg, logger, pgPool, mediasvc.NewMinioStorage(minioClient, cfg.Storage), mediaOpts...)
		pub := notify.NewPostgresPublisher(pgPool)
		if queueProducer != nil {
			pub.SetQueueProducer(queueProducer)
		}
		orderSvc = ordersvc.NewService(cfg, logger, pgPool, pub)
		paymentSvc = paymentsvc.NewService(cfg, logger, pgPool)
		shippingSvc = shippingsvc.NewService(cfg, logger, pgPool)
		sellerSvc = sellersvc.NewService(cfg, logger, pgPool)
		if queueProducer != nil {
			orderSvc.SetQueueProducer(queueProducer)
			paymentSvc.SetQueueProducer(queueProducer)
			sellerSvc.SetQueueProducer(queueProducer)
		}

		var osClient *platformsearch.Client
		if c, err := platformsearch.NewClient(cfg.Search); err == nil {
			osClient = c
		} else {
			logger.Warn("opensearch_client_init_failed", zap.Error(err))
		}
		searchSvc = searchsvc.NewService(cfg, logger, pgPool, osClient)
		promoSvc = promotionsvc.NewService(cfg, logger, pgPool)
		reviewSvc = reviewsvc.NewService(cfg, logger, pgPool, pub)
		notifySvc = notificationsvc.NewService(cfg, logger, pgPool, pub)
		chatSvc = chatsvc.NewService(cfg, logger, pgPool, pub)
	}

	router := httpapi.NewRouter(cfg, logger, readiness, httpapi.RouterOptions{
		RedisClient:         redisClient,
		OpenAPIYAML:         openapi.YAML,
		ExposeMetrics:       cfg.HTTP.MetricsAddr == "",
		TracingEnabled:      cfg.Observability.TracingEnabled,
		AuthService:         authSvc,
		AdminService:        adminSvc,
		CatalogService:      catalogSvc,
		MediaService:        mediaSvc,
		OrderService:        orderSvc,
		SellerService:       sellerSvc,
		PaymentService:      paymentSvc,
		ShippingService:     shippingSvc,
		SearchService:       searchSvc,
		PromotionsService:   promoSvc,
		ReviewService:       reviewSvc,
		NotificationService: notifySvc,
		ChatService:         chatSvc,
	})

	mainServer := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	var metricsServer *http.Server
	if cfg.HTTP.MetricsAddr != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		metricsServer = &http.Server{
			Addr:    cfg.HTTP.MetricsAddr,
			Handler: mux,
		}
		go func() {
			logger.Info("metrics_server_starting", zap.String("addr", cfg.HTTP.MetricsAddr))
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("metrics_server_failed", zap.Error(err))
			}
		}()
	}

	go func() {
		logger.Info("api_server_starting", zap.String("addr", cfg.HTTP.Addr))
		if err := mainServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("api_server_failed", zap.Error(err))
		}
	}()

	waitForShutdown(mainServer, metricsServer, shutdownTracer, cfg.HTTP.ShutdownTimeout, logger)
}

func waitForShutdown(
	mainServer *http.Server,
	metricsServer *http.Server,
	shutdownTracer func(context.Context) error,
	timeout time.Duration,
	logger *zap.Logger,
) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("api_server_stopping")

	if metricsServer != nil {
		if err := metricsServer.Shutdown(ctx); err != nil {
			logger.Error("metrics_server_shutdown_failed", zap.Error(err))
		}
	}

	if err := mainServer.Shutdown(ctx); err != nil {
		logger.Error("api_server_shutdown_failed", zap.Error(err))
	}

	tracerCtx, tracerCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer tracerCancel()
	if err := shutdownTracer(tracerCtx); err != nil {
		logger.Error("tracing_shutdown_failed", zap.Error(err))
	}
}

func valueOrDefault(value string, fallback string) string {
	if value != "" && value != "dev" && value != "local" {
		return value
	}
	return fallback
}
