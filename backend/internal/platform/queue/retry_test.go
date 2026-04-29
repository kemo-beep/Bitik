package queue

import (
	"testing"
	"time"
)

func TestBackoffForAttempt(t *testing.T) {
	def := jobDefinition("jobs.test.v1", "jobs.test")
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 30 * time.Second},
		{2, 2 * time.Minute},
		{3, 10 * time.Minute},
		{4, 30 * time.Minute},
		{5, 1 * time.Hour},
		{6, 3 * time.Hour},
		{7, 6 * time.Hour},
		{8, 12 * time.Hour},
		{12, 12 * time.Hour},
	}
	for _, tc := range cases {
		if got := backoffForAttempt(def, tc.attempt); got != tc.want {
			t.Fatalf("attempt %d backoff got %s want %s", tc.attempt, got, tc.want)
		}
	}
}
