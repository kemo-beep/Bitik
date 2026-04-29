package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/platform/cache"
	"github.com/bitik/backend/internal/platform/db"
	"github.com/bitik/backend/internal/platform/observability"
	"github.com/bitik/backend/internal/platform/queue"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := observability.NewLogger(cfg.Observability)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdownTracer, err := observability.InitTracing(ctx, cfg.Observability, cfg.App.Name+"-worker", cfg.App.Version, cfg.App.Environment)
	if err != nil {
		logger.Fatal("worker.tracing_init_failed", zap.Error(err))
	}

	var redisClient *redis.Client
	if client, err := cache.Connect(ctx, cfg.Redis); err != nil {
		logger.Warn("worker.redis_unavailable", zap.Error(err))
	} else {
		redisClient = client
		defer redisClient.Close()
	}

	pgPool, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("worker.database_connect_failed", zap.Error(err))
	}
	defer pgPool.Close()

	broker, err := queue.NewBroker(cfg.RabbitMQ)
	if err != nil {
		logger.Fatal("worker.rabbitmq_connect_failed", zap.Error(err))
	}
	defer func() {
		if err := broker.Close(); err != nil {
			logger.Warn("worker.rabbitmq_close_failed", zap.Error(err))
		}
	}()
	if err := broker.SetupTopology(); err != nil {
		logger.Fatal("worker.rabbitmq_setup_failed", zap.Error(err))
	}

	app := NewWorkerApp(cfg, logger, pgPool, redisClient, broker)
	if err := app.Start(ctx); err != nil {
		logger.Fatal("worker.start_failed", zap.Error(err))
	}
	metricsSrv := startMetricsServer(cfg, logger)

	logger.Info("worker_ready",
		zap.String("environment", cfg.App.Environment),
		zap.String("version", cfg.App.Version),
	)

	waitForShutdown(cancel, logger)
	if metricsSrv != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = metricsSrv.Shutdown(shutdownCtx)
		shutdownCancel()
	}
	tracerCtx, tracerCancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = shutdownTracer(tracerCtx)
	tracerCancel()
	app.Stop()
	logger.Info("worker_stopped")
}

func waitForShutdown(cancel context.CancelFunc, logger *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("worker_stopping")
	cancel()
	time.Sleep(300 * time.Millisecond)
}

func startMetricsServer(cfg config.Config, logger *zap.Logger) *http.Server {
	if cfg.HTTP.MetricsAddr == "" {
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: cfg.HTTP.MetricsAddr, Handler: mux}
	go func() {
		logger.Info("worker.metrics_server_starting", zap.String("addr", cfg.HTTP.MetricsAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("worker.metrics_server_failed", zap.Error(err))
		}
	}()
	return srv
}
