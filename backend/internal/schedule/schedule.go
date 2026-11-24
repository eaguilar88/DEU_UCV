// Package schedule has the helpers shared by the long-running daily services (cmd/notifier,
// cmd/backup), which wake up once a day at a fixed local time.
package schedule

import (
	"fmt"
	"time"
)

// NextRun returns the first time after now that is hour:minute in loc.
func NextRun(now time.Time, hour, minute int, loc *time.Location) time.Time {
	local := now.In(loc)
	next := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// ParseDaily parses a time of day in HH:MM format and a time zone name.
func ParseDaily(runAt, timeZone string) (hour, minute int, loc *time.Location, err error) {
	loc, err = time.LoadLocation(timeZone)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("loading time zone %q: %w", timeZone, err)
	}
	t, err := time.Parse("15:04", runAt)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("parsing run time %q: %w", runAt, err)
	}
	return t.Hour(), t.Minute(), loc, nil
}
