// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package di

import (
	"github.com/go-logr/logr"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/clock"
	"github.com/int128/scheduled-scaler/pkg/infrastructure/controller"
	"github.com/int128/scheduled-scaler/pkg/repositories/scheduledpodscaler"
	"github.com/int128/scheduled-scaler/pkg/usecases/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Injectors from di.go:

func NewController(logger logr.Logger, clockInterface clock.Interface, clientClient client.Client) controller.Interface {
	repository := &scheduledpodscaler.Repository{
		Client: clientClient,
	}
	reconcileReconcile := &reconcile.Reconcile{
		Log:                          logger,
		Clock:                        clockInterface,
		ScheduledPodScalerRepository: repository,
	}
	controllerController := &controller.Controller{
		Log:     logger,
		UseCase: reconcileReconcile,
	}
	return controllerController
}
