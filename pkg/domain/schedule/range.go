package schedule

import "time"

// Range represents a time range.
type Range interface {
	IsActive(now time.Time) bool
	NextEdge(now time.Time) time.Time
}
