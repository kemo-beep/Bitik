package notify

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/bitik/backend/internal/platform/queue"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgChannel = "bitik_events_v1"

type PostgresPublisher struct {
	pool *pgxpool.Pool

	local *InProcessPublisher
	queue *queue.Producer

	once sync.Once
}

func NewPostgresPublisher(pool *pgxpool.Pool) *PostgresPublisher {
	p := &PostgresPublisher{
		pool:  pool,
		local: NewInProcessPublisher(),
	}
	p.once.Do(func() {
		go p.listenLoop()
	})
	return p
}

func (p *PostgresPublisher) Subscribe(userID string, buffer int) Subscription {
	return p.local.Subscribe(userID, buffer)
}

func (p *PostgresPublisher) Publish(ctx context.Context, evt Event) {
	// Always fan out locally first.
	p.local.Publish(ctx, evt)

	if p.pool == nil {
		return
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return
	}
	// Best-effort cross-process fanout.
	_, _ = p.pool.Exec(ctx, "SELECT pg_notify($1, $2)", pgChannel, string(payload))
	if p.queue != nil && evt.Type == EventNotificationCreated {
		_ = p.queue.PublishFanout(ctx, stringValue(evt.Data["notification_id"]))
	}
}

func (p *PostgresPublisher) SetQueueProducer(producer *queue.Producer) {
	p.queue = producer
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (p *PostgresPublisher) listenLoop() {
	if p.pool == nil {
		return
	}
	for {
		ctx := context.Background()
		conn, err := p.pool.Acquire(ctx)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		_, err = conn.Exec(ctx, "LISTEN "+pgChannel)
		if err != nil {
			conn.Release()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for {
			n, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				break
			}
			var evt Event
			if err := json.Unmarshal([]byte(n.Payload), &evt); err != nil {
				continue
			}
			p.local.Publish(ctx, evt)
		}
		conn.Release()
		time.Sleep(250 * time.Millisecond)
	}
}
