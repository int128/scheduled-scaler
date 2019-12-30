package reconcile

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
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
		return nil, &retryableError{
			error: xerrors.Errorf("could not get the ScheduledPodScaler: %w", err),
		}
	}

	var output Output
	now := r.Clock.Now()
	for _, rule := range scheduledPodScaler.Spec.Rules {
		var rng schedule.Range
		var err error
		switch {
		case rule.Daily != nil:
			rng, err = schedule.NewDailyRange(rule.Daily.StartTime, rule.Daily.EndTime)
			if err != nil {
				return nil, xerrors.Errorf("invalid daily syntax: %w", err)
			}
		}

		next := rng.NextEdge(now).Sub(now)
		if output.NextReconcileAfter == 0 || next < output.NextReconcileAfter {
			output.NextReconcileAfter = next
		}
	}

	//TODO: update the status

	return &output, nil
}
