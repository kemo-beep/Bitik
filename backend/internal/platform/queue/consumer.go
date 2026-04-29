package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(context.Context, Envelope) error

type Consumer struct {
	broker *Broker
	guard  *Guard
}

func NewConsumer(b *Broker) *Consumer {
	return &Consumer{broker: b}
}

func (c *Consumer) WithGuard(g *Guard) *Consumer {
	c.guard = g
	return c
}

func (c *Consumer) Consume(ctx context.Context, jobType JobType, handler HandlerFunc) error {
	if c == nil || c.broker == nil || c.broker.ch == nil {
		return errors.New("consumer broker is not initialized")
	}
	def, ok := JobDefinitions[jobType]
	if !ok {
		return fmt.Errorf("unknown job type: %s", jobType)
	}
	if def.Prefetch > 0 {
		if err := c.broker.ch.Qos(def.Prefetch, 0, false); err != nil {
			return err
		}
	}
	msgs, err := c.broker.ch.Consume(def.QueueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	workers := def.Concurrency
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					consumedTotal.WithLabelValues(string(jobType)).Inc()
					var env Envelope
					if err := json.Unmarshal(msg.Body, &env); err != nil {
						_ = msg.Nack(false, false)
						nackedTotal.WithLabelValues(string(jobType)).Inc()
						dlqTotal.WithLabelValues(string(jobType)).Inc()
						continue
					}
					if err := env.Validate(); err != nil {
						_ = msg.Nack(false, false)
						nackedTotal.WithLabelValues(string(jobType)).Inc()
						dlqTotal.WithLabelValues(string(jobType)).Inc()
						continue
					}
					var runID *pgtype.UUID
					var executionID *pgtype.UUID
					if c.guard != nil {
						run, err := c.guard.Claim(ctx, env)
						if err == ErrDuplicateDone {
							_ = msg.Ack(false)
							ackedTotal.WithLabelValues(string(jobType)).Inc()
							continue
						}
						if err != nil {
							_ = msg.Nack(false, true)
							nackedTotal.WithLabelValues(string(jobType)).Inc()
							continue
						}
						v := run.ID
						runID = &v
						exec, err := c.guard.StartExecution(ctx, *runID, env.MessageID, int32(env.Attempt))
						if err == nil {
							ev := exec.ID
							executionID = &ev
						}
					}
					started := time.Now()
					inFlight.WithLabelValues(string(jobType)).Inc()
					if err := handler(ctx, env); err != nil {
						inFlight.WithLabelValues(string(jobType)).Dec()
						observeHandler(jobType, started, err)
						if runID != nil {
							_ = c.guard.MarkFailed(ctx, *runID, err.Error())
						}
						if executionID != nil {
							_ = c.guard.MarkExecutionFailed(ctx, *executionID, err.Error())
						}
						if env.Attempt >= MaxAttempts {
							_ = msg.Nack(false, false)
							nackedTotal.WithLabelValues(string(jobType)).Inc()
							dlqTotal.WithLabelValues(string(jobType)).Inc()
							continue
						}
						_ = c.publishRetry(ctx, def, env)
						_ = msg.Ack(false)
						retriedTotal.WithLabelValues(string(jobType)).Inc()
						continue
					}
					inFlight.WithLabelValues(string(jobType)).Dec()
					observeHandler(jobType, started, nil)
					if runID != nil {
						_ = c.guard.MarkDone(ctx, *runID)
					}
					if executionID != nil {
						_ = c.guard.MarkExecutionDone(ctx, *executionID)
					}
					_ = msg.Ack(false)
					ackedTotal.WithLabelValues(string(jobType)).Inc()
				}
			}
		}()
	}
	return nil
}

func (c *Consumer) publishRetry(ctx context.Context, def JobDefinition, env Envelope) error {
	env.Attempt++
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return c.broker.ch.PublishWithContext(ctx, ExchangeJobsRetry, def.RoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    env.MessageID,
		Expiration:   strconv.FormatInt(backoffForAttempt(def, env.Attempt).Milliseconds(), 10),
		Body:         body,
	})
}

func backoffForAttempt(def JobDefinition, attempt int) time.Duration {
	if len(def.RetryBackoff) == 0 {
		return 30 * time.Second
	}
	idx := attempt - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(def.RetryBackoff) {
		return def.RetryBackoff[len(def.RetryBackoff)-1]
	}
	return def.RetryBackoff[idx]
}
