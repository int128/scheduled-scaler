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
	"github.com/int128/scheduled-scaler/pkg/repositories/deployment/mock_deployment"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler/mock_scheduledpodscaler"
	kapps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
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
		GetByName(gomock.Not(nil), types.NamespacedName{
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

	mockDeploymentRepository := mock_deployment.NewMockInterface(ctrl)
	mockDeploymentRepository.EXPECT().
		FindBySelectors(gomock.Not(nil), map[string]string{"app": "server1"}).
		Return(&kapps.DeploymentList{
			Items: []kapps.Deployment{
				{
					Spec: kapps.DeploymentSpec{
						Replicas: pointer.Int32Ptr(3),
					},
				},
			},
		}, nil)
	mockDeploymentRepository.EXPECT().
		Scale(gomock.Not(nil), &kapps.Deployment{
			Spec: kapps.DeploymentSpec{
				Replicas: pointer.Int32Ptr(3),
			},
		}, int32(5))

	tc := testingClock(time.Date(2019, 12, 1, 15, 0, 0, 0, time.UTC))
	r := Reconcile{
		Log:                          testingLogr.TestLogger{T: t},
		Clock:                        tc,
		ScheduledPodScalerRepository: mockScheduledPodScalerRepository,
		DeploymentRepository:         mockDeploymentRepository,
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
