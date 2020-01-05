package reconcile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/clock"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler"
	"golang.org/x/xerrors"
	kapps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	//DeploymentRepository         deployment.Interface

	//FIXME: do not expose the client to the use-case
	Client client.Client
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
	scheduledPodScaler, err := r.ScheduledPodScalerRepository.Get(ctx, in.Target)
	if err != nil {
		//TODO: do not retry for the not found error
		return nil, &retryableError{
			error: xerrors.Errorf("could not get the ScheduledPodScaler: %w", err),
		}
	}

	now := r.Clock.Now()
	desiredScaleSpec := scheduledPodScaler.Spec.ComputeDesiredScaleSpec(now)
	scheduledPodScaler.Status.NextReconcileTime = scheduledPodScaler.Spec.FindNextReconcileTime(now)

	//TODO: extract repositories
	selectors := scheduledPodScaler.Spec.ScaleTarget.Selectors
	var l kapps.DeploymentList
	if err := r.Client.List(ctx, &l, client.MatchingLabels(selectors)); err != nil {
		return nil, xerrors.Errorf("could not list the Deployments: %w", err)
	}
	r.Log.Info(fmt.Sprintf("found %d deployments", len(l.Items)))
	for _, deployment := range l.Items {
		var replicas int32
		if deployment.Spec.Replicas != nil {
			replicas = *(deployment.Spec.Replicas)
		}
		r.Log.Info(fmt.Sprintf("comparing the replicas: got=%d, want=%d", replicas, desiredScaleSpec.Replicas))
		if desiredScaleSpec.Replicas != replicas {
			mergePatch, err := json.Marshal(map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": desiredScaleSpec.Replicas,
				},
			})
			if err != nil {
				return nil, xerrors.Errorf("could not create a merge patch: %w", err)
			}
			patch := client.ConstantPatch(types.MergePatchType, mergePatch)
			r.Log.Info("applying the patch to the Deployment", "mergePatch", string(mergePatch))
			if err := r.Client.Patch(ctx, &deployment, patch); err != nil {
				return nil, xerrors.Errorf("could not patch the Deployment: %w", err)
			}
		}
	}

	if err := r.ScheduledPodScalerRepository.UpdateStatus(ctx, scheduledPodScaler); err != nil {
		return nil, &retryableError{
			error: xerrors.Errorf("could not update the status of ScheduledPodScaler: %w", err),
		}
	}
	return &Output{NextReconcileAfter: scheduledPodScaler.Status.NextReconcileTime.Sub(now)}, nil
}
