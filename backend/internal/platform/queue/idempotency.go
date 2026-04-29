package queue

import (
	"context"
	"errors"

	workerstore "github.com/bitik/backend/internal/store/worker"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrDuplicateDone = errors.New("job already completed")

type Guard struct {
	q *workerstore.Queries
}

func NewGuard(q *workerstore.Queries) *Guard {
	return &Guard{q: q}
}

func (g *Guard) Claim(ctx context.Context, evt Envelope) (workerstore.WorkerJobRun, error) {
	row, err := g.q.ClaimWorkerJobRun(ctx, workerstore.ClaimWorkerJobRunParams{
		JobType:   string(evt.JobType),
		DedupeKey: evt.DedupeKey,
		Payload:   evt.Payload,
	})
	if err != nil {
		return workerstore.WorkerJobRun{}, err
	}
	if row.Status == "done" {
		return row, ErrDuplicateDone
	}
	return row, nil
}

func (g *Guard) MarkDone(ctx context.Context, runID pgtype.UUID) error {
	return g.q.MarkWorkerJobDone(ctx, runID)
}

func (g *Guard) MarkFailed(ctx context.Context, runID pgtype.UUID, reason string) error {
	return g.q.MarkWorkerJobFailed(ctx, workerstore.MarkWorkerJobFailedParams{
		LastError: text(reason),
		ID:        runID,
	})
}

func (g *Guard) StartExecution(ctx context.Context, runID pgtype.UUID, messageID string, attempt int32) (workerstore.WorkerJobExecution, error) {
	return g.q.CreateWorkerJobExecution(ctx, workerstore.CreateWorkerJobExecutionParams{
		JobRunID:  runID,
		MessageID: messageID,
		Attempt:   attempt,
	})
}

func (g *Guard) MarkExecutionDone(ctx context.Context, executionID pgtype.UUID) error {
	return g.q.MarkWorkerJobExecutionDone(ctx, executionID)
}

func (g *Guard) MarkExecutionFailed(ctx context.Context, executionID pgtype.UUID, reason string) error {
	return g.q.MarkWorkerJobExecutionFailed(ctx, workerstore.MarkWorkerJobExecutionFailedParams{
		ID:    executionID,
		Error: text(reason),
	})
}

func text(v string) pgtype.Text {
	return pgtype.Text{String: v, Valid: v != ""}
}
