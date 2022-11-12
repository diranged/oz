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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/api/v1alpha1"
)

// ExecAccessRequestReconciler reconciles a ExecAccessRequest object
type ExecAccessRequestReconciler struct {
	// Pass in the common functions from our BaseController
	*BaseReconciler
}

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests/finalizers,verbs=update

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execacesstemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExecAccessRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ExecAccessRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := r.GetLogger(ctx)
	logger.Info("Starting reconcile loop")

	// TODO: If this resource is deleted, then we need to find all AccessRequests pointing to it,
	// and delete them as well.
	request, err := r.GetResource(ctx, req)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to find ExecAccessTemplate %s, perhaps deleted.", req))
		return ctrl.Result{}, nil
	}

	// Verify the resource is valid
	r.Verify(ctx, request)
	if err := r.Status().Update(ctx, request); err != nil {
		logger.Error(err, "Failed to update ExecAccessRequest status")
		return ctrl.Result{}, err
	}

	// If this resource already has a status.podName field set, then we respect that no matter what.
	// We never mutate the pod that this access request was originally created for. Otherwise, pick
	// a Pod and populate that status field.
	if request.Status.PodName != "" {
		logger.Info(fmt.Sprintf("Pod already selected - %s", request.Status.PodName))
	} else {
		// If the user supplied their own Pod, then get that Pod back to make sure it exists. Otherwise,
		// randomly select a pod.
		tmpl, _ := r.getTemplate(ctx, req.Namespace, request.Spec.TemplateName)
		pod, _ := tmpl.GetRandomPod(r.Client, ctx)
		if pod != nil {
			request.Status.PodName = pod.Name
			if err := r.Status().Update(ctx, request); err != nil {
				logger.Error(err, "Failed to update ExecAccessRequest status")
				return ctrl.Result{}, err
			}

			// Let's re-fetch the Custom Resource after update the status so that we have the latest
			// state of the resource on the cluster and we will avoid raise the issue "the object has
			// been modified, please apply your changes to the latest version and try again" which would
			// re-trigger the reconciliation if we try to update it again in the following operations
			if err := r.Get(ctx, req.NamespacedName, request); err != nil {
				logger.Error(err, "Failed to re-fetch ExecAccessRequest")
				return ctrl.Result{}, err
			}

		}
	}

	return ctrl.Result{}, nil
}

// GetResource returns back an ExecAccessRequest resource matching the request supplied to the reconciler loop, or
// returns back an error.
func (r *ExecAccessRequestReconciler) GetResource(ctx context.Context, req ctrl.Request) (*api.ExecAccessRequest, error) {
	tmpl := &api.ExecAccessRequest{}
	err := r.Get(ctx, req.NamespacedName, tmpl)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Verify provides validation for a ExecAccessRequest resource. If at any point the validation fails, the status
// of that resource is updated to indicate that is degraded and cannot be used.
func (r *ExecAccessRequestReconciler) Verify(ctx context.Context, req *api.ExecAccessRequest) error {
	statusType := api.RequestValidated

	if _, err := r.getTemplate(ctx, req.Namespace, req.Spec.TemplateName); err != nil {
		meta.SetStatusCondition(&req.Status.Conditions, metav1.Condition{
			Type:               statusType,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: 0,
			LastTransitionTime: metav1.Time{},
			Reason:             string(metav1.StatusReasonNotFound),
			Message:            fmt.Sprintf("Error: %s", err),
		})
		return err
	}

	meta.SetStatusCondition(&req.Status.Conditions, metav1.Condition{
		Type:               statusType,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: 0,
		LastTransitionTime: metav1.Time{},
		Reason:             api.RequestValidatedSuccess,
		Message:            "Validation successful",
	})

	return nil
}

func (r *ExecAccessRequestReconciler) getTemplate(ctx context.Context, namespace string, name string) (*api.ExecAccessTemplate, error) {
	logger := r.GetLogger(ctx)
	found := &api.ExecAccessTemplate{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, found)

	if err != nil {
		logger.Info("Unable to find ExecAccessTemplate")
		return nil, err
	}

	return found, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExecAccessRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ExecAccessRequest{}).
		Complete(r)
}
