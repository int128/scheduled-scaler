/*
Copyright 2019 Hidetake Iwata.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/int128/scheduled-scaler/pkg/usecases/reconcile"
	kapps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scheduledscalingv1 "github.com/int128/scheduled-scaler/api/v1"
)

// ScheduledPodScalerReconciler reconciles a ScheduledPodScaler object
type ScheduledPodScalerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=scheduledscaling.int128.github.io,resources=scheduledpodscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scheduledscaling.int128.github.io,resources=scheduledpodscalers/status,verbs=get;update;patch

func (r *ScheduledPodScalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("scheduledpodscaler", req.NamespacedName)

	var scheduledPodScalerList scheduledscalingv1.ScheduledPodScalerList
	if err := r.List(ctx, &scheduledPodScalerList, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "could not list the ScheduledScalingList")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info(fmt.Sprintf("found %d ScheduledPodScalers", len(scheduledPodScalerList.Items)))

	var tp timeProvider
	reconcileUseCase := reconcile.Reconcile{
		Log:          log,
		TimeProvider: &tp,
	}
	input := reconcile.Input{ScheduledPodScalerList: scheduledPodScalerList}
	output, err := reconcileUseCase.Do(input)
	if err != nil {
		log.Error(err, "could not determine reconciling", "input", input)
	}

	for _, cmd := range output.ScaleCommands {
		if err := scaleDeployments(ctx, r.Client, log, cmd.ScaleTargetRef, cmd.Spec); err != nil {
			return ctrl.Result{}, err
		}
	}

	if !output.NextReconcileTime.IsZero() {
		requeueAfter := output.NextReconcileTime.Sub(tp.Now())
		log.Info(fmt.Sprintf("requeue after %s", requeueAfter))
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: requeueAfter,
		}, nil
	}
	return ctrl.Result{}, nil
}

type timeProvider struct{}

func (t timeProvider) Now() time.Time {
	return time.Now()
}

//TODO: move to use-case
func scaleDeployments(ctx context.Context, c client.Client, log logr.Logger, ref scheduledscalingv1.ScaleTargetRef, spec scheduledscalingv1.Spec) error {
	labels := client.MatchingLabels(ref.Selectors)
	log.Info("finding deployments by labels", "labels", labels)
	var deploymentList kapps.DeploymentList
	if err := c.List(ctx, &deploymentList, labels); err != nil {
		log.Error(err, "could not list the DeploymentList")
		return client.IgnoreNotFound(err)
	}
	log.Info(fmt.Sprintf("found %d Deployments", len(deploymentList.Items)))

	for _, deploymentItem := range deploymentList.Items {
		log.Info(fmt.Sprintf("scaling the Deployment %s:%s from %d pod(s) to %d pod(s)",
			deploymentItem.Namespace, deploymentItem.Name,
			*deploymentItem.Spec.Replicas,
			spec.Replicas))
		replicas := int32(spec.Replicas)
		deploymentItem.Spec.Replicas = &replicas
		if err := c.Update(ctx, &deploymentItem); err != nil {
			log.Error(err, fmt.Sprintf("could not update the Deployment %s:%s", deploymentItem.Namespace, deploymentItem.Name))
			return err
		}
	}
	return nil
}

func (r *ScheduledPodScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scheduledscalingv1.ScheduledPodScaler{}).
		Complete(r)
}
