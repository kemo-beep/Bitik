-- name: ClaimWorkerJobRun :one
INSERT INTO worker_job_runs (job_type, dedupe_key, status, payload, attempts)
VALUES (
  sqlc.arg('job_type'),
  sqlc.arg('dedupe_key'),
  'processing',
  COALESCE(sqlc.narg('payload')::jsonb, '{}'::jsonb),
  1
)
ON CONFLICT (job_type, dedupe_key) DO UPDATE
SET attempts = worker_job_runs.attempts + 1,
    status = CASE
      WHEN worker_job_runs.status = 'done' THEN worker_job_runs.status
      ELSE 'processing'
    END,
    payload = COALESCE(sqlc.narg('payload')::jsonb, worker_job_runs.payload),
    updated_at = now()
RETURNING id, job_type, dedupe_key, status, attempts, last_error, payload, started_at, completed_at, created_at, updated_at;

-- name: GetWorkerJobRun :one
SELECT id, job_type, dedupe_key, status, attempts, last_error, payload, started_at, completed_at, created_at, updated_at
FROM worker_job_runs
WHERE job_type = $1 AND dedupe_key = $2;

-- name: MarkWorkerJobDone :exec
UPDATE worker_job_runs
SET status = 'done',
    completed_at = now(),
    last_error = NULL,
    updated_at = now()
WHERE id = $1;

-- name: MarkWorkerJobFailed :exec
UPDATE worker_job_runs
SET status = 'failed',
    last_error = sqlc.narg('last_error'),
    updated_at = now()
WHERE id = sqlc.arg('id');

-- name: CreateWorkerJobExecution :one
INSERT INTO worker_job_executions (job_run_id, message_id, attempt, status)
VALUES (sqlc.arg('job_run_id'), sqlc.arg('message_id'), sqlc.arg('attempt'), 'running')
ON CONFLICT (job_run_id, attempt) DO UPDATE
SET message_id = EXCLUDED.message_id,
    status = 'running',
    error = NULL,
    started_at = now(),
    finished_at = NULL
RETURNING id, job_run_id, message_id, attempt, status, error, started_at, finished_at, created_at;

-- name: MarkWorkerJobExecutionDone :exec
UPDATE worker_job_executions
SET status = 'done',
    error = NULL,
    finished_at = now()
WHERE id = $1;

-- name: MarkWorkerJobExecutionFailed :exec
UPDATE worker_job_executions
SET status = 'failed',
    error = sqlc.narg('error'),
    finished_at = now()
WHERE id = sqlc.arg('id');
