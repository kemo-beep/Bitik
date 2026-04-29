package queue

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Envelope struct {
	MessageID     string          `json:"message_id"`
	JobType       JobType         `json:"job_type"`
	SchemaVersion string          `json:"schema_version"`
	CreatedAt     time.Time       `json:"created_at"`
	TraceID       string          `json:"trace_id,omitempty"`
	DedupeKey     string          `json:"dedupe_key"`
	Attempt       int             `json:"attempt"`
	Payload       json.RawMessage `json:"payload"`
}

func (e Envelope) Validate() error {
	if e.MessageID == "" {
		return errors.New("message_id is required")
	}
	if e.JobType == "" {
		return errors.New("job_type is required")
	}
	if e.DedupeKey == "" {
		return errors.New("dedupe_key is required")
	}
	if e.SchemaVersion == "" {
		return errors.New("schema_version is required")
	}
	return nil
}

func NewEnvelope(jobType JobType, dedupeKey string, payload any, traceID string) (Envelope, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		MessageID:     uuid.NewString(),
		JobType:       jobType,
		SchemaVersion: "v1",
		CreatedAt:     time.Now().UTC(),
		TraceID:       traceID,
		DedupeKey:     dedupeKey,
		Attempt:       1,
		Payload:       body,
	}, nil
}
