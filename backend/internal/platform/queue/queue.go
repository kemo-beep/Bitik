package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bitik/backend/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(cfg config.RabbitMQConfig) (*amqp.Connection, error) {
	return amqp.Dial(cfg.URL)
}

type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewBroker(cfg config.RabbitMQConfig) (*Broker, error) {
	conn, err := Connect(cfg)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Broker{conn: conn, ch: ch}, nil
}

func (b *Broker) Close() error {
	var out error
	if b.ch != nil {
		if err := b.ch.Close(); err != nil {
			out = err
		}
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil && out == nil {
			out = err
		}
	}
	return out
}

func (b *Broker) SetupTopology() error {
	if b == nil || b.ch == nil {
		return errors.New("queue broker is not initialized")
	}
	if err := b.ch.ExchangeDeclare(ExchangeJobs, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := b.ch.ExchangeDeclare(ExchangeJobsRetry, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := b.ch.ExchangeDeclare(ExchangeJobsDLQ, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	for _, def := range JobDefinitions {
		if err := b.declareJobQueue(def); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broker) declareJobQueue(def JobDefinition) error {
	if b == nil || b.ch == nil {
		return errors.New("queue broker is not initialized")
	}
	qArgs := amqp.Table{
		"x-dead-letter-exchange":    ExchangeJobsDLQ,
		"x-dead-letter-routing-key": def.RoutingKey,
	}
	if _, err := b.ch.QueueDeclare(def.QueueName, true, false, false, false, qArgs); err != nil {
		return err
	}
	if err := b.ch.QueueBind(def.QueueName, def.RoutingKey, ExchangeJobs, false, nil); err != nil {
		return err
	}

	if _, err := b.ch.QueueDeclare(def.DLQName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := b.ch.QueueBind(def.DLQName, def.RoutingKey, ExchangeJobsDLQ, false, nil); err != nil {
		return err
	}

	if _, err := b.ch.QueueDeclare(def.RetryQueueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    ExchangeJobs,
		"x-dead-letter-routing-key": def.RoutingKey,
	}); err != nil {
		return err
	}
	return b.ch.QueueBind(def.RetryQueueName, def.RoutingKey, ExchangeJobsRetry, false, nil)
}

func (b *Broker) Publish(ctx context.Context, evt Envelope) error {
	if b == nil || b.ch == nil {
		return errors.New("queue broker is not initialized")
	}
	def, ok := JobDefinitions[evt.JobType]
	if !ok {
		return fmt.Errorf("unknown job type: %s", evt.JobType)
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	if err := b.ch.PublishWithContext(ctx, ExchangeJobs, def.RoutingKey, false, false, amqp.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		MessageId:     evt.MessageID,
		CorrelationId: evt.TraceID,
		Type:          string(evt.JobType),
		Timestamp:     time.Now().UTC(),
		Body:          body,
	}); err != nil {
		return err
	}
	publishedTotal.WithLabelValues(string(evt.JobType)).Inc()
	return nil
}

func (b *Broker) QueueDepth(queueName string) (int, error) {
	if b == nil || b.ch == nil {
		return 0, errors.New("queue broker is not initialized")
	}
	state, err := b.ch.QueueInspect(queueName)
	if err != nil {
		return 0, err
	}
	return state.Messages, nil
}

func (b *Broker) Purge(queueName string) error {
	if b == nil || b.ch == nil {
		return errors.New("queue broker is not initialized")
	}
	_, err := b.ch.QueuePurge(queueName, false)
	return err
}
