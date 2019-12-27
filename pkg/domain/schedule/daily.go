package schedule

import (
	"time"

	"golang.org/x/xerrors"
)

// NewDailyRange returns a DailyRange with the given range.
// If endTime < startTime, it treats the endTime as the next day.
// For example, if startTime=23:00:00 and endTime=01:00:00 are given, endTime will be 25:00:00.
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
	if e.Before(s) {
		e = e.AddDate(0, 0, 1)
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

// NextEdge returns the earlier of the next StartTime or EndTime.
func (d *DailyRange) NextEdge(now time.Time) time.Time {
	day := truncateDay(now)
	s := day.Add(d.StartTime)
	if s.Before(now) {
		s = s.AddDate(0, 0, 1)
	}
	e := day.Add(d.EndTime)
	if e.Before(now) {
		e = e.AddDate(0, 0, 1)
	}
	if s.Before(e) {
		return s
	}
	return e
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
