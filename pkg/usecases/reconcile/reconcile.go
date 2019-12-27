package reconcile

import (
	"time"

	"github.com/go-logr/logr"
	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
	"golang.org/x/xerrors"
)

type TimeProvider interface {
	Now() time.Time
}

type Reconcile struct {
	Log          logr.Logger
	TimeProvider TimeProvider
}

type Input struct {
	ScheduledPodScalerList scheduledscalingv1.ScheduledPodScalerList
}

type ScaleCommand struct {
	ScaleTargetRef scheduledscalingv1.ScaleTargetRef
	Spec           scheduledscalingv1.Spec
}

type Output struct {
	ScaleCommands     []*ScaleCommand
	NextReconcileTime time.Time
}

func (r *Reconcile) Do(in Input) (*Output, error) {
	var output Output
	for _, scheduledPodScaler := range in.ScheduledPodScalerList.Items {
		result, err := r.process(scheduledPodScaler)
		if err != nil {
			return nil, xerrors.Errorf("could not process the ScheduledPodScaler: %w", err)
		}

		if output.NextReconcileTime.IsZero() || result.nextReconcileTime.Before(output.NextReconcileTime) {
			output.NextReconcileTime = result.nextReconcileTime
		}
		if result.scaleCommand != nil {
			output.ScaleCommands = append(output.ScaleCommands, result.scaleCommand)
		}
	}
	return &output, nil
}

type processResult struct {
	scaleCommand      *ScaleCommand
	nextReconcileTime time.Time
}

func (r *Reconcile) process(scheduledPodScaler scheduledscalingv1.ScheduledPodScaler) (*processResult, error) {
	now := r.TimeProvider.Now()
	var result processResult

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

		if result.scaleCommand == nil && rng.IsActive(now) {
			result.scaleCommand = &ScaleCommand{
				ScaleTargetRef: scheduledPodScaler.Spec.ScaleTargetRef,
				Spec:           rule.Spec,
			}
		}
		edge := rng.NextEdge(now)
		if result.nextReconcileTime.IsZero() || edge.Before(result.nextReconcileTime) {
			result.nextReconcileTime = edge
		}
	}
	return &result, nil
}
