/*
Copyright 2022 Matt Wise.

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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers/builders"
)

// AccessTemplateReconciler reconciles a AccessTemplate object
type AccessTemplateReconciler struct {
	*OzTemplateReconciler
}

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=accesstemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=accesstemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=accesstemplates/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AccessTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AccessTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting reconcile loop")

	// SETUP
	r.setReconciliationInterval()

	// Get the ExecAccessTemplate resource if it exists. If not, we bail out quietly.
	//
	// TODO: If this resource is deleted, then we need to find all AccessRequests pointing to it,
	// and delete them as well.
	logger.Info("Verifying AccessTemplate exists")
	resource, err := api.GetAccessTemplate(ctx, r.Client, req.Name, req.Namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to find AccessTemplate %s, perhaps deleted.", req))
		return ctrl.Result{}, nil
	}

	// Create an ExecAccessBuilder resource for this particular template, which we'll use to then verify the resource.
	builder := &builders.AccessBuilder{
		BaseBuilder: &builders.BaseBuilder{
			Client:   r.Client,
			Ctx:      ctx,
			Scheme:   r.Scheme,
			Template: resource,
		},
	}

	// VERIFICATION: Make sure that the TargetRef is valid and points to an active controller
	err = r.VerifyTargetRef(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// VERIFICATION: Make sure the DefaultDuration and MaxDuration settings are valid
	err = r.VerifyMiscSettings(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// // TODO:
	// // VERIFICATION: Ensure that the allowedGroups match valid group name strings

	// FINAL: Set Status.Ready state
	err = r.setReadyStatus(ctx, resource)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Duration(r.ReconcililationInterval * int(time.Minute))}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccessTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.AccessTemplate{}).
		Complete(r)
}
