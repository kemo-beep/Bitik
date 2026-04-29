package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/mediasvc"
	"github.com/bitik/backend/internal/notificationsvc"
	"github.com/bitik/backend/internal/ordersvc"
	"github.com/bitik/backend/internal/paymentsvc"
	"github.com/bitik/backend/internal/platform/queue"
	platformsearch "github.com/bitik/backend/internal/platform/search"
	platformstorage "github.com/bitik/backend/internal/platform/storage"
	"github.com/bitik/backend/internal/searchsvc"
	"github.com/bitik/backend/internal/shippingsvc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type WorkerApp struct {
	cfg    config.Config
	log    *zap.Logger
	pool   *pgxpool.Pool
	redis  *redis.Client
	broker *queue.Broker

	orders        *ordersvc.Service
	payments      *paymentsvc.Service
	shipping      *shippingsvc.Service
	search        *searchsvc.Service
	notifications *notificationsvc.Service
	media         *mediasvc.Service
	scheduler     *queue.Scheduler
	adapter       MessageAdapter
}

func NewWorkerApp(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, redis *redis.Client, broker *queue.Broker) *WorkerApp {
	app := &WorkerApp{
		cfg:    cfg,
		log:    logger,
		pool:   pool,
		redis:  redis,
		broker: broker,
	}
	app.initServices()
	return app
}

func (w *WorkerApp) initServices() {
	w.orders = ordersvc.NewService(w.cfg, w.log, w.pool, nil)
	w.payments = paymentsvc.NewService(w.cfg, w.log, w.pool)
	w.shipping = shippingsvc.NewService(w.cfg, w.log, w.pool)

	var osClient *platformsearch.Client
	if c, err := platformsearch.NewClient(w.cfg.Search); err == nil {
		osClient = c
	}
	w.search = searchsvc.NewService(w.cfg, w.log, w.pool, osClient)
	w.notifications = notificationsvc.NewService(w.cfg, w.log, w.pool, nil)

	var storageClient *minio.Client
	if c, err := platformstorage.Connect(context.Background(), w.cfg.Storage); err == nil {
		storageClient = c
	}
	w.media = mediasvc.NewService(w.cfg, w.log, w.pool, mediasvc.NewMinioStorage(storageClient, w.cfg.Storage))
	w.adapter = NewDBLogAdapter(w.log, w.pool)
}

func (w *WorkerApp) Start(ctx context.Context) error {
	if err := w.registerConsumers(ctx); err != nil {
		return err
	}
	w.scheduler = queue.NewScheduler()
	w.scheduler.Every(ctx, 1*time.Minute, w.publishScheduled)
	return nil
}

func (w *WorkerApp) Stop() {
	if w.scheduler != nil {
		w.scheduler.Stop()
	}
}

func (w *WorkerApp) publishScheduled(ctx context.Context) {
	locked, err := w.acquireScheduleLock(ctx)
	if err != nil || !locked {
		return
	}
	defer w.releaseScheduleLock(context.Background())
	w.publishWindowJob(ctx, queue.JobExpireCheckout)
	w.publishWindowJob(ctx, queue.JobCancelUnpaidOrders)
	w.publishWindowJob(ctx, queue.JobReleaseExpiredInventory)
	w.publishWindowJob(ctx, queue.JobUpdateShipmentTracking)
	w.publishWindowJob(ctx, queue.JobSettleSellerWallets)
	w.publishWindowJob(ctx, queue.JobProcessPayouts)
	w.publishWindowJob(ctx, queue.JobPaymentConfirmationTimeout)
	w.publishWindowJob(ctx, queue.JobWaveStaleOrderTimeout)
	w.publishWindowJob(ctx, queue.JobGenerateReports)
}

func (w *WorkerApp) publishWindowJob(ctx context.Context, jobType queue.JobType) {
	env, err := queue.NewEnvelope(jobType, fmt.Sprintf("%s:%s", jobType, time.Now().UTC().Format("2006-01-02T15:04")), map[string]any{
		"window_start": time.Now().UTC().Format(time.RFC3339),
	}, "")
	if err != nil {
		return
	}
	_ = w.broker.Publish(ctx, env)
}

func (w *WorkerApp) acquireScheduleLock(ctx context.Context) (bool, error) {
	row := w.pool.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", int64(9010121))
	var ok bool
	if err := row.Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (w *WorkerApp) releaseScheduleLock(ctx context.Context) {
	_, _ = w.pool.Exec(ctx, "SELECT pg_advisory_unlock($1)", int64(9010121))
}
