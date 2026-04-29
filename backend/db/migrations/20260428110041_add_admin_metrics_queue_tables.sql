-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS analytics_event_queue (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_log_id UUID NOT NULL REFERENCES event_logs(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT,
  available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (event_log_id)
);

CREATE INDEX IF NOT EXISTS idx_analytics_event_queue_status_available
ON analytics_event_queue(status, available_at, created_at);

CREATE TRIGGER trg_analytics_event_queue_updated_at
BEFORE UPDATE ON analytics_event_queue
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS admin_metrics_daily (
  metric_date DATE NOT NULL,
  event_name VARCHAR(150) NOT NULL,
  total_count BIGINT NOT NULL DEFAULT 0,
  unique_users BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (metric_date, event_name)
);

CREATE INDEX IF NOT EXISTS idx_admin_metrics_daily_date
ON admin_metrics_daily(metric_date);

CREATE TRIGGER trg_admin_metrics_daily_updated_at
BEFORE UPDATE ON admin_metrics_daily
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admin_metrics_daily;
DROP TABLE IF EXISTS analytics_event_queue;
-- +goose StatementEnd
