package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextRun(t *testing.T) {
	loc := time.FixedZone("VET", -4*60*60)

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before run time runs today",
			now:  time.Date(2026, 10, 1, 7, 0, 0, 0, loc),
			want: time.Date(2026, 10, 1, 8, 0, 0, 0, loc),
		},
		{
			name: "at run time runs tomorrow",
			now:  time.Date(2026, 10, 1, 8, 0, 0, 0, loc),
			want: time.Date(2026, 10, 2, 8, 0, 0, 0, loc),
		},
		{
			name: "after run time runs tomorrow",
			now:  time.Date(2026, 10, 1, 21, 0, 0, 0, loc),
			want: time.Date(2026, 10, 2, 8, 0, 0, 0, loc),
		},
		{
			name: "now in another zone is converted",
			now:  time.Date(2026, 10, 2, 3, 0, 0, 0, time.UTC), // 23:00 on Oct 1 in loc
			want: time.Date(2026, 10, 2, 8, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextRun(tt.now, 8, 0, loc)
			assert.True(t, tt.want.Equal(got), "got %s, want %s", got, tt.want)
		})
	}
}
