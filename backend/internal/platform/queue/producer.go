package queue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Producer struct {
	broker *Broker
}

func NewProducer(b *Broker) *Producer {
	return &Producer{broker: b}
}

func (p *Producer) Enqueue(ctx context.Context, fileID uuid.UUID) error {
	if p == nil || p.broker == nil {
		return nil
	}
	payload := map[string]string{"file_id": fileID.String()}
	env, err := NewEnvelope(JobProcessImage, fmt.Sprintf("image:%s", fileID.String()), payload, "")
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, env)
}

func (p *Producer) PublishFanout(ctx context.Context, notificationID string) error {
	if p == nil || p.broker == nil {
		return nil
	}
	env, err := NewEnvelope(JobNotificationFanout, "fanout:"+notificationID, map[string]string{"notification_id": notificationID}, "")
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, env)
}

func (p *Producer) PublishJob(ctx context.Context, jobType JobType, dedupeKey string, payload any) error {
	if p == nil || p.broker == nil {
		return nil
	}
	env, err := NewEnvelope(jobType, dedupeKey, payload, "")
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, env)
}
