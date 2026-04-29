package paymentsvc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/bitik/backend/internal/config"
)

func TestVerifyWebhookSignature_AcceptReject(t *testing.T) {
	s := &Service{
		cfg: config.Config{
			Payments: config.PaymentsConfig{
				WebhookSecret: "test-secret",
			},
		},
	}

	provider := "wave_manual"
	body := []byte(`{"event":"paid"}`)

	mac := hmac.New(sha256.New, []byte("test-secret:"+provider))
	_, _ = mac.Write(body)
	valid := hex.EncodeToString(mac.Sum(nil))

	if err := s.verifyWebhookSignature(provider, "sha256="+valid, body); err != nil {
		t.Fatalf("expected valid signature, got err: %v", err)
	}
	if err := s.verifyWebhookSignature(provider, "sha256=deadbeef", body); err == nil {
		t.Fatal("expected invalid signature error")
	}
	if err := s.verifyWebhookSignature(provider, "", body); err == nil {
		t.Fatal("expected missing signature error")
	}
}

func TestVerifyWebhookSignature_AllowsWhenSecretUnset(t *testing.T) {
	s := &Service{cfg: config.Config{}}
	if err := s.verifyWebhookSignature("wave_manual", "", []byte(`{}`)); err != nil {
		t.Fatalf("expected nil when secret unset, got %v", err)
	}
}
