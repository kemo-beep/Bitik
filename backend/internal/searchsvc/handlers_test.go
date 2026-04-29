package searchsvc

import "testing"

func TestParseOptionalUUIDStrict(t *testing.T) {
	if _, ok := parseOptionalUUIDStrict(""); !ok {
		t.Fatal("empty UUID should be accepted")
	}
	if _, ok := parseOptionalUUIDStrict("not-a-uuid"); ok {
		t.Fatal("invalid UUID should fail")
	}
	if id, ok := parseOptionalUUIDStrict("f47ac10b-58cc-4372-a567-0e02b2c3d479"); !ok || !id.Valid {
		t.Fatalf("valid UUID should parse, ok=%v valid=%v", ok, id.Valid)
	}
}

func TestParseOptionalInt64(t *testing.T) {
	if _, ok := parseOptionalInt64(""); !ok {
		t.Fatal("empty int should be accepted")
	}
	if _, ok := parseOptionalInt64("-1"); ok {
		t.Fatal("negative int should fail")
	}
	if v, ok := parseOptionalInt64("42"); !ok || !v.Valid || v.Int64 != 42 {
		t.Fatalf("expected 42, ok=%v valid=%v value=%d", ok, v.Valid, v.Int64)
	}
}
