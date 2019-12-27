package schedule

import (
	"time"

	"golang.org/x/xerrors"
)

// NewDailyRange returns a DailyRange with the given range.
func NewDailyRange(startTime, endTime string) (*DailyRange, error) {
	b := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	s, err := time.Parse("15:04:05", startTime)
	if err != nil {
		return nil, xerrors.Errorf("could not parse the startTime: %w", err)
	}
	e, err := time.Parse("15:04:05", endTime)
	if err != nil {
		return nil, xerrors.Errorf("could not parse the endTime: %w", err)
	}
	return &DailyRange{
		StartTime: s.Sub(b),
		EndTime:   e.Sub(b),
	}, nil
}

// DailyRange represents a daily schedule.
type DailyRange struct {
	StartTime time.Duration
	EndTime   time.Duration
}

// IsActive returns true if t is in the range.
// This function depends on the timezone of t.
func (d *DailyRange) IsActive(t time.Time) bool {
	since := t.Sub(truncateDay(t))
	return d.StartTime < since && since < d.EndTime
}

// NextStartTime returns the next StartTime.
// If the StartTime is over, it returns the StartTime of tomorrow.
func (d *DailyRange) NextStartTime(t time.Time) time.Time {
	next := truncateDay(t).Add(d.StartTime)
	if next.Before(t) {
		return next.AddDate(0, 0, 1)
	}
	return next
}

// NextEndTime returns the next EndTime.
// If the EndTime is over, it returns the EndTime of tomorrow.
func (d *DailyRange) NextEndTime(t time.Time) time.Time {
	next := truncateDay(t).Add(d.EndTime)
	if next.Before(t) {
		return next.AddDate(0, 0, 1)
	}
	return next
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
