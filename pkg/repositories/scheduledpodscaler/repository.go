package scheduledpodscaler

import (
	"context"

	"github.com/google/wire"
	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
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
	Get(ctx context.Context, name types.NamespacedName) (*scheduledscalingv1.ScheduledPodScaler, error)
	UpdateStatus(ctx context.Context, o scheduledscalingv1.ScheduledPodScaler) error
}

type Repository struct {
	Client client.Client
}

func (r *Repository) Get(ctx context.Context, name types.NamespacedName) (*scheduledscalingv1.ScheduledPodScaler, error) {
	var o scheduledscalingv1.ScheduledPodScaler
	if err := r.Client.Get(ctx, name, &o); err != nil {
		return nil, xerrors.Errorf("could not get the item: %w", err)
	}
	return &o, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, o scheduledscalingv1.ScheduledPodScaler) error {
	if err := r.Client.Status().Update(ctx, &o); err != nil {
		return xerrors.Errorf("could not update the status: %w", err)
	}
	return nil
}
