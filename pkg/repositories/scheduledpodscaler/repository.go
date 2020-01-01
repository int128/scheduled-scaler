package scheduledpodscaler

import (
	"context"
	"time"

	"github.com/google/wire"
	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
	"github.com/int128/scheduled-scaler/pkg/domain/scheduledpodscaler"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Repository)),
	wire.Struct(new(Repository), "*"),
)

//go:generate mockgen -destination mock_scheduledpodscaler/mock_scheduledpodscaler.go github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler Interface

type Interface interface {
	Get(ctx context.Context, name types.NamespacedName) (*scheduledpodscaler.ScheduledPodScaler, error)
	UpdateStatus(ctx context.Context, s *scheduledpodscaler.ScheduledPodScaler) error
}

type Repository struct {
	Client client.Client
}

func (r *Repository) Get(ctx context.Context, name types.NamespacedName) (*scheduledpodscaler.ScheduledPodScaler, error) {
	var o scheduledscalingv1.ScheduledPodScaler
	if err := r.Client.Get(ctx, name, &o); err != nil {
		return nil, xerrors.Errorf("could not get the item: %w", err)
	}
	var s scheduledpodscaler.ScheduledPodScaler
	s.TypeMeta, s.ObjectMeta = o.TypeMeta, o.ObjectMeta

	for _, rule := range o.Spec.Rules {
		var rng schedule.Range
		var err error
		switch {
		case rule.Daily != nil:
			rng, err = schedule.NewDailyRange(rule.Daily.StartTime, rule.Daily.EndTime)
			if err != nil {
				return nil, xerrors.Errorf("invalid daily syntax: %w", err)
			}
		}
		s.Spec.ScaleRules = append(s.Spec.ScaleRules, scheduledpodscaler.ScaleRule{
			Range: rng,
			ScaleSpec: scheduledpodscaler.ScaleSpec{
				Replicas: rule.Spec.Replicas,
			},
		})
	}
	s.Spec.ScaleTarget.Selectors = o.Spec.ScaleTargetRef.Selectors

	t, err := time.Parse(time.RFC3339, o.Status.NextReconcileTime)
	if err != nil {
		return nil, xerrors.Errorf("could not parse Status.NextReconcileTime: %w", err)
	}
	s.Status.NextReconcileTime = t

	return &s, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, s *scheduledpodscaler.ScheduledPodScaler) error {
	var o scheduledscalingv1.ScheduledPodScaler
	o.TypeMeta, o.ObjectMeta = s.TypeMeta, s.ObjectMeta

	o.Status.NextReconcileTime = s.Status.NextReconcileTime.Format(time.RFC3339)

	if err := r.Client.Status().Update(ctx, &o); err != nil {
		return xerrors.Errorf("could not update the status: %w", err)
	}
	return nil
}
