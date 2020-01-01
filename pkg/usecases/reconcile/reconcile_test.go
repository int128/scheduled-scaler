package reconcile

import (
	"context"
	"testing"
	"time"

	testingLogr "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
	"github.com/int128/scheduled-scaler/pkg/domain/scheduledpodscaler"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler/mock_scheduledpodscaler"
	"k8s.io/apimachinery/pkg/types"
)

type testingClock time.Time

func (t testingClock) Now() time.Time {
	return time.Time(t)
}

func TestReconcile_Do(t *testing.T) {
	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockScheduledPodScalerRepository := mock_scheduledpodscaler.NewMockInterface(ctrl)
	mockScheduledPodScalerRepository.EXPECT().
		Get(gomock.Not(nil), types.NamespacedName{
			Namespace: "fixture",
			Name:      "example1",
		}).
		Return(&scheduledpodscaler.ScheduledPodScaler{
			Spec: scheduledpodscaler.Spec{
				ScaleTarget: scheduledpodscaler.ScaleTarget{
					Selectors: map[string]string{
						"app": "server1",
					},
				},
				ScaleRules: []scheduledpodscaler.ScaleRule{
					{
						Range: &schedule.DailyRange{
							StartTime: 12 * time.Hour,
							EndTime:   19 * time.Hour,
						},
						ScaleSpec: scheduledpodscaler.ScaleSpec{
							Replicas: 5,
						},
					},
				},
			},
		}, nil)
	mockScheduledPodScalerRepository.EXPECT().
		UpdateStatus(gomock.Not(nil), &scheduledpodscaler.ScheduledPodScaler{
			Spec: scheduledpodscaler.Spec{
				ScaleTarget: scheduledpodscaler.ScaleTarget{
					Selectors: map[string]string{
						"app": "server1",
					},
				},
				ScaleRules: []scheduledpodscaler.ScaleRule{
					{
						Range: &schedule.DailyRange{
							StartTime: 12 * time.Hour,
							EndTime:   19 * time.Hour,
						},
						ScaleSpec: scheduledpodscaler.ScaleSpec{
							Replicas: 5,
						},
					},
				},
			},
			Status: scheduledpodscaler.Status{
				NextReconcileTime: time.Date(2019, 12, 1, 19, 0, 0, 0, time.UTC),
			},
		})
	tc := testingClock(time.Date(2019, 12, 1, 15, 0, 0, 0, time.UTC))
	r := Reconcile{
		Log:                          testingLogr.TestLogger{T: t},
		Clock:                        tc,
		ScheduledPodScalerRepository: mockScheduledPodScalerRepository,
	}

	input := Input{
		Target: types.NamespacedName{
			Namespace: "fixture",
			Name:      "example1",
		},
	}
	got, err := r.Do(ctx, input)
	if err != nil {
		t.Fatalf("Do error: %+v", err)
	}
	want := &Output{
		NextReconcileAfter: 4 * time.Hour,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want, +got):\n%s", diff)
	}
}
