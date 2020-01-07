package reconcile

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/clock"
	"github.com/int128/scheduled-scaler/pkg/repositories/deployment"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Reconcile)),
	wire.Struct(new(Reconcile), "*"),
)

type Interface interface {
	Do(ctx context.Context, in Input) (*Output, error)
}

type Reconcile struct {
	Log                          logr.Logger
	Clock                        clock.Interface
	ScheduledPodScalerRepository scheduledpodscaler.Interface
	DeploymentRepository         deployment.Interface
}

type Input struct {
	Target types.NamespacedName
}

type Output struct {
	NextReconcileAfter time.Duration
}

type retryableError struct {
	error
}

func (e *retryableError) IsRetryable() bool {
	return true
}

func (r *Reconcile) Do(ctx context.Context, in Input) (*Output, error) {
	scheduledPodScaler, err := r.ScheduledPodScalerRepository.GetByName(ctx, in.Target)
	if err != nil {
		//TODO: do not retry for the not found error
		return nil, &retryableError{
			error: xerrors.Errorf("could not get the ScheduledPodScaler: %w", err),
		}
	}

	selectors := scheduledPodScaler.Spec.ScaleTarget.Selectors
	deploymentList, err := r.DeploymentRepository.FindBySelectors(ctx, selectors)
	if err != nil {
		return nil, xerrors.Errorf("could not find the deployments: %w", err)
	}
	r.Log.Info(fmt.Sprintf("found %d deployments", len(deploymentList.Items)), "selectors", selectors)

	now := r.Clock.Now()
	desiredScaleSpec := scheduledPodScaler.Spec.ComputeDesiredScaleSpec(now)
	for _, deploymentItem := range deploymentList.Items {
		currentReplicas := pointer.Int32PtrDerefOr(deploymentItem.Spec.Replicas, 0)
		r.Log.Info("comparing the replicas", "current", currentReplicas, "desired", desiredScaleSpec.Replicas)
		if currentReplicas != desiredScaleSpec.Replicas {
			r.Log.Info("applying the patch to the deployment", "replicas", currentReplicas)
			if err := r.DeploymentRepository.Scale(ctx, &deploymentItem, desiredScaleSpec.Replicas); err != nil {
				return nil, xerrors.Errorf("could not scale the deployment: %w", err)
			}
		}
	}

	scheduledPodScaler.Status.NextReconcileTime = scheduledPodScaler.Spec.FindNextReconcileTime(now)
	if err := r.ScheduledPodScalerRepository.UpdateStatus(ctx, scheduledPodScaler); err != nil {
		return nil, &retryableError{
			error: xerrors.Errorf("could not update the status of ScheduledPodScaler: %w", err),
		}
	}
	return &Output{NextReconcileAfter: scheduledPodScaler.Status.NextReconcileTime.Sub(now)}, nil
}
