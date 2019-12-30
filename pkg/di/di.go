//go:generate wire
//+build wireinject

package di

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/clock"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/controller"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler"
	"github.com/int128/scheduled-scaler/pkg/usecases/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewController(logr.Logger, clock.Interface, client.Client) controller.Interface {
	wire.Build(
		// usecases
		reconcile.Set,

		// repositories
		scheduledpodscaler.Set,

		// infrastructure
		controller.Set,
	)
	return nil
}
