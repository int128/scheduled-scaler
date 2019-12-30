package reconcile

import (
	"context"
	"testing"
	"time"

	testingLogr "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
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
		Return(&scheduledscalingv1.ScheduledPodScaler{
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
		}, nil)
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
