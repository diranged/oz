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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers/builders"
)

// ExecAccessRequestReconciler reconciles a ExecAccessRequest object
type ExecAccessRequestReconciler struct {
	// Pass in the common functions from our BaseController
	*OzRequestReconciler
}

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests/finalizers,verbs=update

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execacesstemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// https://kubernetes.io/docs/concepts/security/rbac-good-practices/#escalate-verb
//
// We leverage the escalate verb here because we don't specifically want or need the Oz controller
// pods to have Exec/Debug privileges on pods, but we want them to be able to grant those privileges
// to users.
//
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;bind;escalate
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

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
	logger := log.FromContext(ctx)
	logger.Info("Starting reconcile loop")

	// SETUP
	r.SetReconciliationInterval()

	// First make sure we use the ApiReader (non-cached) client to go and figure out if the resource exists or not. If
	// it doesn't come back, we exit out beacuse it is likely the object has been deleted and we no longer need to
	// worry about it.
	logger.Info("Verifying ExecAccessRequest exists")
	resource, err := api.GetExecAccessRequest(r.Client, ctx, req.Name, req.Namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to find ExecAccessRequest %s, perhaps deleted.", req.Name))
		return ctrl.Result{}, nil
	}

	// VERIFICATION: Make sure the Target TemplateName field points to a valid Template
	tmpl, err := r.getTargetTemplate(ctx, resource)
	if err != nil {
		return ctrl.Result{}, err
	}

	// OWNER UPDATE: Update the ExecAccessRequest OwnerRef to the TargetTemplate.
	//
	// Ensure that if the TargetTemplate is ever deleted, that all of the AccessRequests are
	// also deleted, which will cascade down and delete any roles/bindings/etc.
	//
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	//
	// TODO: BUGFIX< THIS IS NOT PUSHING THE UPDATE TO K8S
	if err := ctrl.SetControllerReference(tmpl, resource, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Create an AccessBuilder resource for this particular template, which we'll use to then verify the resource.
	builder := &builders.ExecAccessBuilder{
		BaseBuilder: &builders.BaseBuilder{
			// Boilerplate
			Client: r.Client,
			Ctx:    ctx,
			Scheme: r.Scheme,
			// Ours
			ApiReader: r.ApiReader,
			Request:   resource,
			Template:  tmpl,
		},
		Request:  resource,
		Template: tmpl,
	}

	// VERIFICATION: Verifies the requested duration
	err = r.VerifyDuration(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// VERIFICATION: Handle whether or not the access is expired at this point! If so, delete it.
	if expired, err := r.IsAccessExpired(builder); err != nil {
		return ctrl.Result{}, err
	} else if expired {
		return ctrl.Result{}, nil
	}

	// VERIFICATION: Make sure all of the access resources are built properly. On any failure,
	// set up a 30 second delay before the next reconciliation attempt.
	_, err = r.VerifyAccessResources(builder)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, err
	}

	// FINAL: Set Status.Ready state
	err = r.SetReadyStatus(ctx, resource)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Exit Reconciliation Loop
	logger.Info("Ending reconcile loop")

	// Finally, requeue to re-reconcile again in the future
	return ctrl.Result{RequeueAfter: time.Duration(r.ReconcililationInterval * int(time.Minute))}, nil
}

// getTargetTemplate is used to both verify that the desired Spec.TemplateName field actually exists in the cluster,
// and to return that populated object back to the reconciler loop. The ConditionTargetTemplateExists condition is
// updated with the status.
//
// Returns:
//   - Pointer to the api.ExecAccessTemplate (or nil)
//   - An "error" only if the UpdateCondition function fails
func (r *ExecAccessRequestReconciler) getTargetTemplate(ctx context.Context, req *api.ExecAccessRequest) (*api.ExecAccessTemplate, error) {
	logger := r.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Verifying that Target Template %s still exists...", req.Spec.TemplateName))

	if tmpl, err := api.GetExecAccessTemplate(r.Client, ctx, req.Spec.TemplateName, req.Namespace); err != nil {
		// On failure: Update the condition, and return.
		return nil, r.UpdateCondition(
			ctx, req, ConditionTargetTemplateExists, metav1.ConditionFalse,
			string(metav1.StatusReasonNotFound), fmt.Sprintf("Error: %s", err))

	} else {
		return tmpl, r.UpdateCondition(
			ctx, req, ConditionTargetTemplateExists, metav1.ConditionTrue, string(metav1.StatusSuccess),
			"Found Target Template")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExecAccessRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Provide a searchable index in the cached kubernetes client for "metadata.name" - the pod name.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.Pod{}, fieldSelectorMetadataName, func(rawObj client.Object) []string {
		// grab the job object, extract the name...
		pod := rawObj.(*v1.Pod)
		name := pod.GetName()
		return []string{name}
	}); err != nil {
		return err
	}

	// Provide a searchable index in the cached kubernetes client for "status.phase", allowing us to
	// search for Running Pods.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.Pod{}, fieldSelectorStatusPhase, func(rawObj client.Object) []string {
		// grab the job object, extract the phase...
		pod := rawObj.(*v1.Pod)
		phase := string(pod.Status.Phase)
		return []string{phase}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ExecAccessRequest{}).
		Complete(r)
}
