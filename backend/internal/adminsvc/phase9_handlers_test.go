package adminsvc

import "testing"

func TestReportStatusFromCaseStatus(t *testing.T) {
	cases := map[string]string{
		"under_review": "under_review",
		"resolved":    "resolved",
		"dismissed":   "dismissed",
		"unknown_val": "",
	}
	for in, want := range cases {
		if got := reportStatusFromCaseStatus(in); got != want {
			t.Fatalf("status %s -> %s, want %s", in, got, want)
		}
	}
}
