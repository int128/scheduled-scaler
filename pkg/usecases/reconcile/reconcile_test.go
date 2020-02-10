package reconcile

import (
	"context"
	"fmt"
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

	t.Run("ScaleDeployment", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		scheduledPodScaler1 := scheduledpodscaler.ScheduledPodScaler{
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
						Timezone: time.UTC,
						ScaleSpec: scheduledpodscaler.ScaleSpec{
							Replicas: 5,
						},
					},
				},
			},
		}
		mockScheduledPodScalerRepository := mock_scheduledpodscaler.NewMockInterface(ctrl)
		mockScheduledPodScalerRepository.EXPECT().
			GetByName(gomock.Not(nil), types.NamespacedName{
				Namespace: "fixture",
				Name:      "example1",
			}).
			Return(&scheduledPodScaler1, nil)
		mockScheduledPodScalerRepository.EXPECT().
			UpdateStatus(gomock.Not(nil), &scheduledpodscaler.ScheduledPodScaler{
				Spec: scheduledPodScaler1.Spec,
				Status: scheduledpodscaler.Status{
					NextReconcileTime: time.Date(2019, 12, 1, 19, 0, 0, 0, time.UTC),
				},
			})

		deployment1 := kapps.Deployment{
			Spec: kapps.DeploymentSpec{
				Replicas: pointer.Int32Ptr(3),
			},
		}
		mockDeploymentRepository := mock_deployment.NewMockInterface(ctrl)
		mockDeploymentRepository.EXPECT().
			FindBySelectors(gomock.Not(nil), map[string]string{"app": "server1"}).
			Return(&kapps.DeploymentList{
				Items: []kapps.Deployment{deployment1},
			}, nil)
		mockDeploymentRepository.EXPECT().
			Scale(gomock.Not(nil), &deployment1, int32(5))

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
	})

	t.Run("NotScaleDeployment", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		scheduledPodScaler1 := scheduledpodscaler.ScheduledPodScaler{
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
						Timezone: time.UTC,
						ScaleSpec: scheduledpodscaler.ScaleSpec{
							Replicas: 5,
						},
					},
				},
			},
		}
		mockScheduledPodScalerRepository := mock_scheduledpodscaler.NewMockInterface(ctrl)
		mockScheduledPodScalerRepository.EXPECT().
			GetByName(gomock.Not(nil), types.NamespacedName{
				Namespace: "fixture",
				Name:      "example1",
			}).
			Return(&scheduledPodScaler1, nil)
		mockScheduledPodScalerRepository.EXPECT().
			UpdateStatus(gomock.Not(nil), &scheduledpodscaler.ScheduledPodScaler{
				Spec: scheduledPodScaler1.Spec,
				Status: scheduledpodscaler.Status{
					NextReconcileTime: time.Date(2019, 12, 1, 19, 0, 0, 0, time.UTC),
				},
			})

		deployment1 := kapps.Deployment{
			Spec: kapps.DeploymentSpec{
				Replicas: pointer.Int32Ptr(5),
			},
		}
		mockDeploymentRepository := mock_deployment.NewMockInterface(ctrl)
		mockDeploymentRepository.EXPECT().
			FindBySelectors(gomock.Not(nil), map[string]string{"app": "server1"}).
			Return(&kapps.DeploymentList{
				Items: []kapps.Deployment{deployment1},
			}, nil)

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
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("ScheduledPodScalerNotFound", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockScheduledPodScalerRepository := mock_scheduledpodscaler.NewMockInterface(ctrl)
			mockScheduledPodScalerRepository.EXPECT().
				GetByName(gomock.Not(nil), types.NamespacedName{
					Namespace: "fixture",
					Name:      "example1",
				}).
				Return(nil, &aError{
					error:     fmt.Errorf("not found error"),
					temporary: true,
					notFound:  true,
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
				NextReconcileAfter: 0,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want, +got):\n%s", diff)
			}
		})
	})
}

type aError struct {
	error
	temporary bool
	notFound  bool
}

func (err *aError) IsTemporary() bool {
	return err.temporary
}

func (err *aError) IsNotFound() bool {
	return err.notFound
}
