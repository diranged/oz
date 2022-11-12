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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/api/v1alpha1"
)

// ExecAccessTemplateReconciler reconciles a ExecAccessTemplate object
type ExecAccessTemplateReconciler struct {
	// Pass in the common functions from our BaseController
	*BaseReconciler
}

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccesstemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccesstemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccesstemplates/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExecAccessTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ExecAccessTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := r.GetLogger(ctx)
	logger.Info("Starting reconcile loop")

	// Get the ExecAccessTemplate resource if it exists. If not, we bail out quietly.
	//
	// TODO: If this resource is deleted, then we need to find all AccessRequests pointing to it,
	// and delete them as well.
	tmpl, err := r.GetResource(ctx, req)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to find ExecAccessTemplate %s, perhaps deleted.", req))
		return ctrl.Result{}, nil
	}

	// Verify the resource is valid
	r.Verify(ctx, tmpl)

	if err := r.Status().Update(ctx, tmpl); err != nil {
		logger.Error(err, "Failed to update ExecAccessTemplate status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

// GetResource returns back an ExecAccessTemplate resource matching the request supplied to the reconciler loop, or
// returns back an error.
func (r *ExecAccessTemplateReconciler) GetResource(ctx context.Context, req ctrl.Request) (*api.ExecAccessTemplate, error) {
	tmpl := &api.ExecAccessTemplate{}
	err := r.Get(ctx, req.NamespacedName, tmpl)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Verify provides validation for a ExecAccessTemplate resource. If at any point the validation fails, the status
// of that resource is updated to indicate that is degraded and cannot be used.
func (r *ExecAccessTemplateReconciler) Verify(ctx context.Context, tmpl *api.ExecAccessTemplate) error {
	statusType := api.TemplateAvailability

	if tmpl.Spec.TargetRef.Kind == api.DeploymentController {
		if _, err := tmpl.GetDeployment(r.Client, ctx); err != nil {
			meta.SetStatusCondition(&tmpl.Status.Conditions, metav1.Condition{
				Type:               statusType,
				Status:             metav1.ConditionUnknown,
				ObservedGeneration: 0,
				LastTransitionTime: metav1.Time{},
				Reason:             string(metav1.StatusReasonNotFound),
				Message:            fmt.Sprintf("Error: %s", err),
			})
			return err
		}
	}

	meta.SetStatusCondition(&tmpl.Status.Conditions, metav1.Condition{
		Type:               statusType,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: 0,
		LastTransitionTime: metav1.Time{},
		Reason:             api.TemplateAvailabilityStatusAvailable,
		Message:            "Verification successful",
	})

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExecAccessTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ExecAccessTemplate{}).
		Complete(r)
}
