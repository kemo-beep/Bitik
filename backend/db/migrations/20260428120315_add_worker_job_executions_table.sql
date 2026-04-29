-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS worker_job_executions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_run_id UUID NOT NULL REFERENCES worker_job_runs(id) ON DELETE CASCADE,
  message_id VARCHAR(100) NOT NULL,
  attempt INT NOT NULL,
  status VARCHAR(30) NOT NULL,
  error TEXT,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_worker_job_executions_run_created
ON worker_job_executions(job_run_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_worker_job_executions_unique_attempt
ON worker_job_executions(job_run_id, attempt);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS worker_job_executions;
-- +goose StatementEnd
