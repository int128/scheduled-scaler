package scheduledpodscaler

import (
	"context"
	"time"

	"github.com/google/wire"
	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
	"github.com/int128/scheduled-scaler/pkg/domain/scheduledpodscaler"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/errors"
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
	GetByName(ctx context.Context, name types.NamespacedName) (*scheduledpodscaler.ScheduledPodScaler, error)
	UpdateStatus(ctx context.Context, s *scheduledpodscaler.ScheduledPodScaler) error
}

type Repository struct {
	Client client.Client
}

// GetByName returns the ScheduledPodScaler of the name.
func (r *Repository) GetByName(ctx context.Context, name types.NamespacedName) (*scheduledpodscaler.ScheduledPodScaler, error) {
	var o scheduledscalingv1.ScheduledPodScaler
	if err := r.Client.Get(ctx, name, &o); err != nil {
		return nil, errors.Wrap(err)
	}
	var s scheduledpodscaler.ScheduledPodScaler
	s.TypeMeta, s.ObjectMeta = o.TypeMeta, o.ObjectMeta

	s.Spec.ScaleTarget.Selectors = o.Spec.ScaleTarget.Selectors

	for _, rule := range o.Spec.ScaleRules {
		var rng schedule.Range
		var err error
		switch {
		case rule.Daily != nil:
			rng, err = schedule.NewDailyRange(rule.Daily.StartTime, rule.Daily.EndTime)
			if err != nil {
				return nil, xerrors.Errorf("invalid daily syntax: %w", err)
			}
		default:
			return nil, xerrors.Errorf("currently only daily is supported")
		}
		s.Spec.ScaleRules = append(s.Spec.ScaleRules, scheduledpodscaler.ScaleRule{
			Range: rng,
			ScaleSpec: scheduledpodscaler.ScaleSpec{
				Replicas: rule.ScaleSpec.Replicas,
			},
		})
	}

	s.Spec.DefaultScaleSpec.Replicas = o.Spec.DefaultScaleSpec.Replicas

	if o.Status.NextReconcileTime != "" {
		t, err := time.Parse(time.RFC3339, o.Status.NextReconcileTime)
		if err != nil {
			return nil, xerrors.Errorf("could not parse Status.NextReconcileTime: %w", err)
		}
		s.Status.NextReconcileTime = t
	}

	return &s, nil
}

// UpdateStatus updates the status. It does not update the spec.
func (r *Repository) UpdateStatus(ctx context.Context, s *scheduledpodscaler.ScheduledPodScaler) error {
	var o scheduledscalingv1.ScheduledPodScaler
	o.TypeMeta, o.ObjectMeta = s.TypeMeta, s.ObjectMeta

	o.Status.NextReconcileTime = s.Status.NextReconcileTime.Format(time.RFC3339)

	if err := r.Client.Status().Update(ctx, &o); err != nil {
		return errors.Wrap(err)
	}
	return nil
}
