package authsvc

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bitik/backend/internal/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestLoginLockoutTransitions(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis run: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{
		cfg: config.Config{
			Auth: config.AuthConfig{
				MaxLoginFailures:     3,
				LoginLockoutDuration: time.Minute,
			},
		},
		redis: rdb,
	}

	ctx := context.Background()
	email := "user@example.com"
	if svc.isLoginLocked(ctx, email) {
		t.Fatal("expected unlocked initially")
	}

	svc.recordLoginFailure(ctx, email)
	svc.recordLoginFailure(ctx, email)
	if svc.isLoginLocked(ctx, email) {
		t.Fatal("should not be locked before threshold")
	}

	svc.recordLoginFailure(ctx, email)
	if !svc.isLoginLocked(ctx, email) {
		t.Fatal("expected locked after threshold reached")
	}

	svc.clearLoginFailures(ctx, email)
	if svc.isLoginLocked(ctx, email) {
		t.Fatal("expected unlocked after clear")
	}
}

func TestOTPVerifyLockoutTransitions(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis run: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{
		cfg: config.Config{
			Auth: config.AuthConfig{
				OTPMaxVerifyFailures: 2,
			},
		},
		redis: rdb,
	}

	ctx := context.Background()
	userID := uuid.New()
	phone := "+233000111222"

	locked := svc.recordOTPVerifyFailure(ctx, userID, phone)
	if locked {
		t.Fatal("first failure should not lock")
	}
	locked = svc.recordOTPVerifyFailure(ctx, userID, phone)
	if !locked {
		t.Fatal("second failure should lock")
	}

	key := svc.otpVerifyFailKey(userID, phone)
	if !mr.Exists(key) {
		t.Fatal("expected otp fail key to exist")
	}

	svc.clearOTPVerifyFailures(ctx, userID, phone)
	if mr.Exists(key) {
		t.Fatal("expected otp fail key cleared")
	}
}
