//go:build integration

package queue

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bitik/backend/internal/config"
	workerstore "github.com/bitik/backend/internal/store/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConsumer_IdempotentDuplicateDelivery(t *testing.T) {
	rabbitURL := os.Getenv("BITIK_TEST_RABBITMQ_URL")
	dbURL := os.Getenv("BITIK_TEST_DATABASE_URL")
	if rabbitURL == "" || dbURL == "" {
		t.Skip("integration env not configured")
	}
	cfg := testRabbitConfig(rabbitURL)
	broker, err := NewBroker(cfg)
	if err != nil {
		t.Fatalf("broker: %v", err)
	}
	defer broker.Close()
	if err := broker.SetupTopology(); err != nil {
		t.Fatalf("topology: %v", err)
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("db pool: %v", err)
	}
	defer pool.Close()
	_, _ = pool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS worker_job_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_type VARCHAR(120) NOT NULL,
  dedupe_key VARCHAR(255) NOT NULL,
  status VARCHAR(30) NOT NULL DEFAULT 'processing',
  attempts INT NOT NULL DEFAULT 1,
  last_error TEXT,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (job_type, dedupe_key)
);`)
	_, _ = pool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS worker_job_executions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_run_id UUID NOT NULL REFERENCES worker_job_runs(id) ON DELETE CASCADE,
  message_id VARCHAR(100) NOT NULL,
  attempt INT NOT NULL,
  status VARCHAR(30) NOT NULL,
  error TEXT,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (job_run_id, attempt)
);`)

	orig := JobDefinitions[JobSendEmail]
	def := orig
	def.Prefetch = 1
	def.Concurrency = 1
	def.RetryBackoff = []time.Duration{20 * time.Millisecond}
	JobDefinitions[JobSendEmail] = def
	defer func() { JobDefinitions[JobSendEmail] = orig }()
	_ = broker.Purge(def.QueueName)
	_ = broker.Purge(def.RetryQueueName)
	_ = broker.Purge(def.DLQName)

	guard := NewGuard(workerstore.New(pool))
	consumer := NewConsumer(broker).WithGuard(guard)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var called int32
	if err := consumer.Consume(ctx, JobSendEmail, func(context.Context, Envelope) error {
		atomic.AddInt32(&called, 1)
		return nil
	}); err != nil {
		t.Fatalf("consume: %v", err)
	}
	env, _ := NewEnvelope(JobSendEmail, "email:test:dup", map[string]string{"x": "1"}, "")
	if err := broker.Publish(context.Background(), env); err != nil {
		t.Fatalf("publish1: %v", err)
	}
	if err := broker.Publish(context.Background(), env); err != nil {
		t.Fatalf("publish2: %v", err)
	}
	time.Sleep(800 * time.Millisecond)
	if got := atomic.LoadInt32(&called); got != 1 {
		t.Fatalf("handler called %d times, want 1", got)
	}
}

func TestConsumer_RetryThenSuccess(t *testing.T) {
	rabbitURL := os.Getenv("BITIK_TEST_RABBITMQ_URL")
	if rabbitURL == "" {
		t.Skip("integration env not configured")
	}
	cfg := testRabbitConfig(rabbitURL)
	broker, err := NewBroker(cfg)
	if err != nil {
		t.Fatalf("broker: %v", err)
	}
	defer broker.Close()
	if err := broker.SetupTopology(); err != nil {
		t.Fatalf("topology: %v", err)
	}
	orig := JobDefinitions[JobSendPush]
	def := orig
	def.Prefetch = 1
	def.Concurrency = 1
	def.RetryBackoff = []time.Duration{30 * time.Millisecond}
	JobDefinitions[JobSendPush] = def
	defer func() { JobDefinitions[JobSendPush] = orig }()
	_ = broker.Purge(def.QueueName)
	_ = broker.Purge(def.RetryQueueName)
	_ = broker.Purge(def.DLQName)

	consumer := NewConsumer(broker)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var called int32
	if err := consumer.Consume(ctx, JobSendPush, func(context.Context, Envelope) error {
		if atomic.AddInt32(&called, 1) == 1 {
			return errors.New("first attempt fails")
		}
		return nil
	}); err != nil {
		t.Fatalf("consume: %v", err)
	}
	env, _ := NewEnvelope(JobSendPush, "push:test:retry", map[string]string{"x": "1"}, "")
	if err := broker.Publish(context.Background(), env); err != nil {
		t.Fatalf("publish: %v", err)
	}
	time.Sleep(1200 * time.Millisecond)
	if got := atomic.LoadInt32(&called); got < 2 {
		t.Fatalf("expected retry call count >=2, got %d", got)
	}
}

func testRabbitConfig(url string) config.RabbitMQConfig {
	return config.RabbitMQConfig{URL: url}
}
