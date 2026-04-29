package queue

import "testing"

func TestNewEnvelopeAndValidate(t *testing.T) {
	env, err := NewEnvelope(JobCancelUnpaidOrders, "cancel_unpaid:window", map[string]any{"window": "now"}, "trace-1")
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}
	if env.Attempt != 1 {
		t.Fatalf("expected attempt=1 got=%d", env.Attempt)
	}
	if err := env.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestEnvelopeValidateRejectsMissingFields(t *testing.T) {
	env := Envelope{}
	if err := env.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
