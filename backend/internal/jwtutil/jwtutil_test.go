package jwtutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSignParseRoundTrip(t *testing.T) {
	secret := "test-secret-at-least-32-bytes-long!!"
	iss := "test-issuer"
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	roles := []string{"buyer", "seller"}

	tok, err := Sign(secret, iss, uid, roles, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Parse(secret, iss, tok)
	if err != nil {
		t.Fatal(err)
	}
	got, err := SubjectUUID(claims)
	if err != nil || got != uid {
		t.Fatalf("subject: got %v err %v", got, err)
	}
	if len(claims.Roles) != len(roles) {
		t.Fatalf("roles: %#v", claims.Roles)
	}
}

func TestParseWrongSecret(t *testing.T) {
	tok, _ := Sign("a", "i", uuid.New(), nil, time.Minute)
	_, err := Parse("b", "i", tok)
	if err == nil {
		t.Fatal("expected error")
	}
}
