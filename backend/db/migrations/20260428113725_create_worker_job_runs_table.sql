-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS worker_job_runs (
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
);

CREATE INDEX IF NOT EXISTS idx_worker_job_runs_status_created
ON worker_job_runs(status, created_at DESC);

CREATE TRIGGER trg_worker_job_runs_updated_at
BEFORE UPDATE ON worker_job_runs
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS worker_job_runs;
-- +goose StatementEnd
