package schedule_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
)

func TestNewDailyRange(t *testing.T) {
	t.Run("NonZero", func(t *testing.T) {
		got, err := schedule.NewDailyRange("01:23:45", "06:07:08")
		if err != nil {
			t.Errorf("NewDailyRange error: %s", err)
		}
		want := &schedule.DailyRange{
			StartTime: parseDuration(t, "1h23m45s"),
			EndTime:   parseDuration(t, "6h7m8s"),
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
		got, err := schedule.NewDailyRange("01:23", "06:07:08")
		if got != nil {
			t.Errorf("NewDailyRange wants nil but %+v", got)
		}
		if err == nil {
			t.Errorf("NewDailyRange wants error but nil")
		}
	})
	t.Run("InvalidEndTime", func(t *testing.T) {
		got, err := schedule.NewDailyRange("01:23:45", "06:07")
		if got != nil {
			t.Errorf("NewDailyRange wants nil but %+v", got)
		}
		if err == nil {
			t.Errorf("NewDailyRange wants error but nil")
		}
	})
}

func TestDailyRange_IsActive(t *testing.T) {
	t.Run("InRange", func(t *testing.T) {
		daily, err := schedule.NewDailyRange("01:23:45", "06:07:08")
		if err != nil {
			t.Fatalf("NewDailyRange error: %s", err)
		}
		got := daily.IsActive(timeProvider(time.Date(2019, 12, 3, 4, 5, 6, 0, time.UTC)))
		if got != true {
			t.Errorf("IsActive wants true but false")
		}
	})
	t.Run("BeforeRange", func(t *testing.T) {
		daily, err := schedule.NewDailyRange("01:23:45", "06:07:08")
		if err != nil {
			t.Fatalf("NewDailyRange error: %s", err)
		}
		got := daily.IsActive(timeProvider(time.Date(2019, 12, 3, 1, 5, 6, 0, time.UTC)))
		if got != false {
			t.Errorf("IsActive wants false but true")
		}
	})
	t.Run("AfterRange", func(t *testing.T) {
		daily, err := schedule.NewDailyRange("01:23:45", "06:07:08")
		if err != nil {
			t.Fatalf("NewDailyRange error: %s", err)
		}
		got := daily.IsActive(timeProvider(time.Date(2019, 12, 3, 7, 5, 6, 0, time.UTC)))
		if got != false {
			t.Errorf("IsActive wants false but true")
		}
	})
}

type timeProvider time.Time

func (t timeProvider) Now() time.Time {
	return time.Time(t)
}

func parseDuration(t *testing.T, s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		t.Errorf("could not parse the duration: %s", err)
	}
	return d
}
