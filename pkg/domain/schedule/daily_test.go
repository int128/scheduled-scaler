package schedule_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
)

func TestNewDailyRange(t *testing.T) {
	t.Run("StartTime<EndTime", func(t *testing.T) {
		got, err := schedule.NewDailyRange("01:23:45", "23:45:06")
		if err != nil {
			t.Errorf("NewDailyRange error: %s", err)
		}
		want := &schedule.DailyRange{
			StartTime: parseDuration(t, "1h23m45s"),
			EndTime:   parseDuration(t, "23h45m6s"),
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("EndTime<StartTime", func(t *testing.T) {
		got, err := schedule.NewDailyRange("23:45:06", "01:23:45")
		if err != nil {
			t.Errorf("NewDailyRange error: %s", err)
		}
		want := &schedule.DailyRange{
			StartTime: parseDuration(t, "23h45m6s"),
			EndTime:   parseDuration(t, "25h23m45s"),
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("Zero", func(t *testing.T) {
		got, err := schedule.NewDailyRange("00:00:00", "00:00:00")
		if err != nil {
			t.Errorf("NewDailyRange error: %s", err)
		}
		want := &schedule.DailyRange{
			StartTime: time.Duration(0),
			EndTime:   time.Duration(0),
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("InvalidStartTime", func(t *testing.T) {
		got, err := schedule.NewDailyRange("01:23", "23:45:06")
		if got != nil {
			t.Errorf("NewDailyRange wants nil but %+v", got)
		}
		if err == nil {
			t.Errorf("NewDailyRange wants error but nil")
		}
	})
	t.Run("InvalidEndTime", func(t *testing.T) {
		got, err := schedule.NewDailyRange("01:23:45", "23:45")
		if got != nil {
			t.Errorf("NewDailyRange wants nil but %+v", got)
		}
		if err == nil {
			t.Errorf("NewDailyRange wants error but nil")
		}
	})
}

func parseDuration(t *testing.T, s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		t.Errorf("could not parse the duration: %s", err)
	}
	return d
}

var timezones = []*time.Location{
	time.UTC,
	mustLocation(time.LoadLocation("Asia/Tokyo")),
	mustLocation(time.LoadLocation("America/Los_Angeles")),
}

func mustLocation(l *time.Location, err error) *time.Location {
	if err != nil {
		panic(err)
	}
	return l
}

func TestDailyRange_IsActive(t *testing.T) {
	tests := func(t *testing.T, tz *time.Location) {
		t.Run("InRange", func(t *testing.T) {
			daily, err := schedule.NewDailyRange("01:23:45", "23:45:06")
			if err != nil {
				t.Fatalf("NewDailyRange error: %s", err)
			}
			got := daily.IsActive(time.Date(2019, 12, 3, 4, 5, 6, 0, tz))
			if got != true {
				t.Errorf("IsActive wants true but false (daily=%+v)", daily)
			}
		})
		t.Run("BeforeRange", func(t *testing.T) {
			daily, err := schedule.NewDailyRange("01:23:45", "23:45:06")
			if err != nil {
				t.Fatalf("NewDailyRange error: %s", err)
			}
			got := daily.IsActive(time.Date(2019, 12, 3, 1, 5, 6, 0, tz))
			if got != false {
				t.Errorf("IsActive wants false but true (daily=%+v)", daily)
			}
		})
		t.Run("AfterRange", func(t *testing.T) {
			daily, err := schedule.NewDailyRange("01:23:45", "23:45:06")
			if err != nil {
				t.Fatalf("NewDailyRange error: %s", err)
			}
			got := daily.IsActive(time.Date(2019, 12, 3, 23, 50, 6, 0, tz))
			if got != false {
				t.Errorf("IsActive wants false but true (daily=%+v)", daily)
			}
		})
	}

	for _, tz := range timezones {
		t.Run(tz.String(), func(t *testing.T) {
			tests(t, tz)
		})
	}
}

func TestDailyRange_NextEdge(t *testing.T) {
	tests := func(t *testing.T, tz *time.Location) {
		t.Run("StartTime<EndTime", func(t *testing.T) {
			// 01:23:45 (today)
			// 23:45:06 (today)
			dailyRange, err := schedule.NewDailyRange("01:23:45", "23:45:06")
			if err != nil {
				t.Fatalf("NewDailyRange error: %s", err)
			}
			t.Run("BeforeStart", func(t *testing.T) {
				now := time.Date(2019, 12, 3, 1, 0, 0, 0, tz)
				got := dailyRange.NextEdge(now)
				want := time.Date(2019, 12, 3, 1, 23, 45, 0, tz)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
			t.Run("AfterStartBeforeEnd", func(t *testing.T) {
				now := time.Date(2019, 12, 3, 2, 0, 0, 0, tz)
				got := dailyRange.NextEdge(now)
				want := time.Date(2019, 12, 3, 23, 45, 6, 0, tz)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
			t.Run("AfterEnd", func(t *testing.T) {
				now := time.Date(2019, 12, 3, 23, 50, 0, 0, tz)
				got := dailyRange.NextEdge(now)
				want := time.Date(2019, 12, 4, 1, 23, 45, 0, tz)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
		})
		t.Run("EndTime<StartTime", func(t *testing.T) {
			// 23:45:06 (today)
			// 01:23:45 (tomorrow +1)
			dailyRange, err := schedule.NewDailyRange("23:45:06", "01:23:45")
			if err != nil {
				t.Fatalf("NewDailyRange error: %s", err)
			}
			t.Run("BeforeEnd", func(t *testing.T) {
				now := time.Date(2019, 12, 3, 1, 0, 0, 0, tz)
				got := dailyRange.NextEdge(now)
				want := time.Date(2019, 12, 3, 23, 45, 6, 0, tz)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
			t.Run("BeforeStart", func(t *testing.T) {
				now := time.Date(2019, 12, 3, 23, 0, 0, 0, tz)
				got := dailyRange.NextEdge(now)
				want := time.Date(2019, 12, 3, 23, 45, 6, 0, tz)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
			t.Run("AfterStart", func(t *testing.T) {
				now := time.Date(2019, 12, 3, 23, 50, 0, 0, tz)
				got := dailyRange.NextEdge(now)
				want := time.Date(2019, 12, 4, 1, 23, 45, 0, tz)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
		})
	}

	for _, tz := range timezones {
		t.Run(tz.String(), func(t *testing.T) {
			tests(t, tz)
		})
	}
}
