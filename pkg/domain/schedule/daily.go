package schedule

import (
	"time"

	"golang.org/x/xerrors"
)

// Now provides the current time.
type Now interface {
	Now() time.Time
}

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

// IsActive returns true if the current time is in the range.
func (d *DailyRange) IsActive(n Now) bool {
	now := n.Now()
	today := now.Truncate(24 * time.Hour)
	todayTime := now.Sub(today)
	return d.StartTime < todayTime && todayTime < d.EndTime
}
