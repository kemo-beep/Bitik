package main

import (
	"context"
	"encoding/json"

	"github.com/bitik/backend/internal/pgxutil"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type MessageAdapter interface {
	Send(ctx context.Context, kind string, payload json.RawMessage) error
}

type DBLogAdapter struct {
	log *zap.Logger
	q   *systemstore.Queries
}

func NewDBLogAdapter(logger *zap.Logger, pool *pgxpool.Pool) *DBLogAdapter {
	return &DBLogAdapter{log: logger, q: systemstore.New(pool)}
}

func (a *DBLogAdapter) Send(ctx context.Context, kind string, payload json.RawMessage) error {
	meta := []byte(`{}`)
	if len(payload) > 0 {
		meta = payload
	}
	uid := pgtype.UUID{}
	var body map[string]any
	_ = json.Unmarshal(payload, &body)
	if raw, ok := body["user_id"].(string); ok {
		if id, err := uuid.Parse(raw); err == nil {
			uid = pgxutil.UUID(id)
		}
	}
	if _, err := a.q.CreateEventLog(ctx, systemstore.CreateEventLogParams{
		UserID:     uid,
		EventName:  "worker." + kind,
		EntityType: text(kind),
		Metadata:   meta,
	}); err != nil {
		return err
	}
	a.log.Info("worker_adapter_logged", zap.String("kind", kind))
	return nil
}

func text(v string) pgtype.Text {
	return pgtype.Text{String: v, Valid: v != ""}
}
