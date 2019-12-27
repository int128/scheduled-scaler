package reconcile_test

import (
	"testing"
	"time"

	testingLogr "github.com/go-logr/logr/testing"
	"github.com/google/go-cmp/cmp"
	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
	"github.com/int128/scheduled-scaler/pkg/usecases/reconcile"
)

type timeProvider time.Time

func (t timeProvider) Now() time.Time {
	return time.Time(t)
}

func TestReconcile_Do(t *testing.T) {
	r := reconcile.Reconcile{
		Log:          testingLogr.TestLogger{T: t},
		TimeProvider: timeProvider(time.Date(2019, 12, 1, 15, 0, 0, 0, time.UTC)),
	}
	input := reconcile.Input{
		ScheduledPodScalerList: scheduledscalingv1.ScheduledPodScalerList{
			Items: []scheduledscalingv1.ScheduledPodScaler{
				{
					Spec: scheduledscalingv1.ScheduledPodScalerSpec{
						ScaleTargetRef: scheduledscalingv1.ScaleTargetRef{
							Selectors: map[string]string{
								"app": "server1",
							},
						},
						Rules: []scheduledscalingv1.Rule{
							{
								Spec: scheduledscalingv1.Spec{
									Replicas: 5,
								},
								Daily: &scheduledscalingv1.Daily{
									StartTime: "12:00:00",
									EndTime:   "19:00:00",
								},
							},
						},
					},
				},
				{
					Spec: scheduledscalingv1.ScheduledPodScalerSpec{
						ScaleTargetRef: scheduledscalingv1.ScaleTargetRef{
							Selectors: map[string]string{
								"app": "server2",
							},
						},
						Rules: []scheduledscalingv1.Rule{
							{
								Spec: scheduledscalingv1.Spec{
									Replicas: 3,
								},
								Daily: &scheduledscalingv1.Daily{
									StartTime: "17:00:00",
									EndTime:   "23:00:00",
								},
							},
						},
					},
				},
			},
		},
	}
	got, err := r.Do(input)
	if err != nil {
		t.Fatalf("Do error: %+v", err)
	}
	want := &reconcile.Output{
		ScaleCommands: []*reconcile.ScaleCommand{
			{
				ScaleTargetRef: scheduledscalingv1.ScaleTargetRef{
					Selectors: map[string]string{
						"app": "server1",
					},
				},
				Spec: scheduledscalingv1.Spec{
					Replicas: 5,
				},
			},
		},
		NextReconcileTime: time.Date(2019, 12, 1, 17, 0, 0, 0, time.UTC),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want, +got):\n%s", diff)
	}
}
