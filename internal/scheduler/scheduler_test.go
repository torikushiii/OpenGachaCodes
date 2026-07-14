package scheduler

import (
	"testing"
	"time"
)

func TestNextBoundary(t *testing.T) {
	location := time.UTC
	tests := []struct {
		now  string
		want string
	}{
		{"2026-07-12T08:17:00Z", "2026-07-12T08:30:00Z"},
		{"2026-07-12T08:30:01Z", "2026-07-12T09:00:00Z"},
		{"2026-07-12T08:59:00Z", "2026-07-12T09:00:00Z"},
		{"2026-07-12T09:00:00Z", "2026-07-12T09:30:00Z"},
		{"2026-07-12T09:30:00Z", "2026-07-12T10:00:00Z"},
	}
	for _, tt := range tests {
		now, _ := time.Parse(time.RFC3339, tt.now)
		want, _ := time.Parse(time.RFC3339, tt.want)
		if got := NextBoundary(now, location); !got.Equal(want) {
			t.Errorf("NextBoundary(%s)=%s, want %s", tt.now, got, want)
		}
	}
}
