package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/int128/scheduled-scaler/pkg/domain/errors"
	"github.com/int128/scheduled-scaler/pkg/usecases/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Controller)),
	wire.Struct(new(Controller), "*"),
)

type Interface interface {
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

type Controller struct {
	Log     logr.Logger
	UseCase reconcile.Interface
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	c.Log.Info("starting reconciliation")
	input := reconcile.Input{
		Target: req.NamespacedName,
	}
	output, err := c.UseCase.Do(ctx, input)
	if err != nil {
		if errors.IsTemporary(err) {
			c.Log.Info("retry reconciliation due to the temporary error", "error", err)
			return ctrl.Result{}, err
		}
		c.Log.Error(err, "permanent error")
	}
	if output.NextReconcileAfter != 0 {
		c.Log.Info(fmt.Sprintf("finished reconciliation and requeue after %s", output.NextReconcileAfter))
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: output.NextReconcileAfter,
		}, nil
	}
	c.Log.Info("finished reconciliation")
	return ctrl.Result{}, nil
}
