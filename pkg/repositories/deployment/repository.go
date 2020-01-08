package deployment

import (
	"context"
	"encoding/json"

	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/errors"
	"golang.org/x/xerrors"
	kapps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Repository)),
	wire.Struct(new(Repository), "*"),
)

//go:generate mockgen -destination mock_deployment/mock_deployment.go github.com/int128/scheduled-scaler/pkg/repositories/deployment Interface

type Interface interface {
	FindBySelectors(ctx context.Context, selectors map[string]string) (*kapps.DeploymentList, error)
	Scale(ctx context.Context, deployment *kapps.Deployment, replicas int32) error
}

type Repository struct {
	Client client.Client
}

// FindBySelectors returns a list of deployments matched to the selectors.
func (r *Repository) FindBySelectors(ctx context.Context, selectors map[string]string) (*kapps.DeploymentList, error) {
	var l kapps.DeploymentList
	if err := r.Client.List(ctx, &l, client.MatchingLabels(selectors)); err != nil {
		return nil, errors.Wrap(err)
	}
	return &l, nil
}

// Scale updates the replicas of the deployment to the given value using the patch method.
func (r *Repository) Scale(ctx context.Context, deployment *kapps.Deployment, replicas int32) error {
	b, err := json.Marshal(&scaleMergePatch{
		Spec: scaleMergePatchSpec{
			Replicas: replicas,
		},
	})
	if err != nil {
		return xerrors.Errorf("could not encode the json: %w", err)
	}
	p := client.ConstantPatch(types.MergePatchType, b)
	if err := r.Client.Patch(ctx, deployment, p); err != nil {
		return errors.Wrap(err)
	}
	deployment.Spec.Replicas = &replicas
	return nil
}

type scaleMergePatch struct {
	Spec scaleMergePatchSpec `json:"spec"`
}

type scaleMergePatchSpec struct {
	Replicas int32 `json:"replicas"`
}
