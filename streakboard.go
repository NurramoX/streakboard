// Package streakboard renders GitHub-style contribution boards from
// arbitrary per-day activity data — habits, vocabulary reviews, anything
// countable.
package streakboard

import "time"

// Entry records activity on one day. Multiple entries for the same
// calendar date are summed.
type Entry struct {
	Date  time.Time // only the calendar date (in Date's own location) is used
	Count int
}
