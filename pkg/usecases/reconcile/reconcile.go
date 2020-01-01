package reconcile

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/clock"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/types"
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
	scheduledPodScaler.Spec.ComputeDesiredScaleSpec(now) //TODO: scale the target
	scheduledPodScaler.Status.NextReconcileTime = scheduledPodScaler.Spec.FindNextReconcileTime(now)

	if err := r.ScheduledPodScalerRepository.UpdateStatus(ctx, scheduledPodScaler); err != nil {
		return nil, &retryableError{
			error: xerrors.Errorf("could not update the status of ScheduledPodScaler: %w", err),
		}
	}
	return &Output{NextReconcileAfter: scheduledPodScaler.Status.NextReconcileTime.Sub(now)}, nil
}
