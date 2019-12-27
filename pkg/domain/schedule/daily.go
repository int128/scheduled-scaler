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
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	since := t.Sub(day)
	return d.StartTime < since && since < d.EndTime
}
