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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	templates "github.com/diranged/oz/api/v1alpha1"
	"github.com/go-logr/logr"
)

// ExecAccessTemplateReconciler reconciles a ExecAccessTemplate object
type ExecAccessTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
}

//+kubebuilder:rbac:groups=templates.wizardoz.co,resources=execaccesstemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=templates.wizardoz.co,resources=execaccesstemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=templates.wizardoz.co,resources=execaccesstemplates/finalizers,verbs=update

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
	r.logger = ctrllog.FromContext(ctx)
	r.logger.Info("Starting reconcile loop")

	// Get the ExecAccessTemplate resource if it exists. If not, we bail out quietly.
	//
	// TODO: If this resource is deleted, then we need to find all AccessRequests pointing to it,
	// and delete them as well.
	tmpl, err := r.GetResource(ctx, req)
	if err != nil {
		r.logger.Info(fmt.Sprintf("Failed to find ExecAccessTemplate %s, perhaps deleted.", req))
		return ctrl.Result{}, nil
	}

	// Verify the resource is valid
	r.Verify(ctx, tmpl)

	if err := r.Status().Update(ctx, tmpl); err != nil {
		r.logger.Error(err, "Failed to update ExecAccessTemplate status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

// GetResource returns back an ExecAccessTemplate resource matching the request supplied to the reconciler loop, or
// returns back an error.
func (r *ExecAccessTemplateReconciler) GetResource(ctx context.Context, req ctrl.Request) (*templates.ExecAccessTemplate, error) {
	tmpl := &templates.ExecAccessTemplate{}
	err := r.Get(ctx, req.NamespacedName, tmpl)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Verify provides validation for a ExecAccessTemplate resource. If at any point the validation fails, the status
// of that resource is updated to indicate that is degraded and cannot be used.
func (r *ExecAccessTemplateReconciler) Verify(ctx context.Context, tmpl *templates.ExecAccessTemplate) error {
	statusType := templates.TemplateAvailability

	if tmpl.Spec.TargetRef.Kind == templates.KindDeployment {
		if _, err := r.getDeployment(ctx, tmpl.Namespace, tmpl.Spec.TargetRef); err != nil {
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
		Reason:             templates.TemplateAvailabilityStatusAvailable,
		Message:            "Verification successful",
	})

	return nil
}

func (r *ExecAccessTemplateReconciler) getDeployment(ctx context.Context, namespace string, targetRef templates.CrossVersionObjectReference) (*appsv1.Deployment, error) {
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      *targetRef.Name,
		Namespace: namespace,
	}, found)

	if err != nil {
		r.logger.Info("Unable to find Deployment")
		return nil, err
	}

	return found, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExecAccessTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&templates.ExecAccessTemplate{}).
		Complete(r)
}
