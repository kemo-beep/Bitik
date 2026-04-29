package authsvc

import "testing"

func TestAppleEmailVerified(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{name: "bool true", in: true, want: true},
		{name: "bool false", in: false, want: false},
		{name: "string true", in: "true", want: true},
		{name: "string false", in: "false", want: false},
		{name: "missing", in: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appleEmailVerified(tt.in); got != tt.want {
				t.Fatalf("appleEmailVerified(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestUserStatusString(t *testing.T) {
	if got := userStatusString("active"); got != "active" {
		t.Fatalf("string status = %q", got)
	}
	if got := userStatusString([]byte("banned")); got != "banned" {
		t.Fatalf("byte status = %q", got)
	}
	if got := userStatusString(123); got != "" {
		t.Fatalf("unknown status = %q", got)
	}
}
