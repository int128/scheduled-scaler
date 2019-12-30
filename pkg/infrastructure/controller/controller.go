package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/usecases/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Controller)),
	wire.Struct(new(Controller), "*"),
)

type UseCaseError interface {
	error
	IsRetryable() bool
}

type Interface interface {
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

type Controller struct {
	Log     logr.Logger
	UseCase reconcile.Interface
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	c.Log.Info("starting reconciling")
	input := reconcile.Input{
		Target: req.NamespacedName,
	}
	output, err := c.UseCase.Do(ctx, input)
	if err != nil {
		c.Log.Error(err, "error while reconciling", "input", input)
		if err, ok := err.(UseCaseError); ok {
			if err.IsRetryable() {
				c.Log.Info("retry reconciling due to error")
				return ctrl.Result{}, err
			}
		}
	}
	if output.NextReconcileAfter != 0 {
		c.Log.Info(fmt.Sprintf("finished reconciling with requeue after %s", output.NextReconcileAfter))
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: output.NextReconcileAfter,
		}, nil
	}
	c.Log.Info("finished reconciling with no requeue")
	return ctrl.Result{}, nil
}
