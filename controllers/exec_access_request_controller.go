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
	if err := ctrl.SetControllerReference(tmpl, resource, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Create an AccessBuilder resource for this particular template, which we'll use to then verify the resource.
	builder := &builders.ExecAccessBuilder{
		BaseBuilder: &builders.BaseBuilder{
			Client:   r.Client,
			Ctx:      ctx,
			Scheme:   r.Scheme,
			Request:  resource,
			Template: tmpl,
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

	// Get or Set the Target Pod Name for the access request. If the Status.TargetPod field is already set, this
	// will simply return that value.
	_, err = r.GetPodName(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// if request.Status.PodName != "" {
	// 	r.UpdateCondition(
	// 		builder.Ctx, builder.Request,
	// 		ConditionTargetPodSelected,
	// 		metav1.ConditionTrue,
	// 		string(metav1.StatusSuccess),
	// 		fmt.Sprintf("Pod %s selected", request.Status.PodName))
	// }

	// // VERIFICATION: Make sure the Target Pod still exists - that it hasn't gone away at some point.
	// r.VerifyTargetPodExists(ctx, request, request.Status.PodName)

	// // RBAC: Make sure the Role exists
	// err = r.CreateOrUpdateRoleStatus(builder)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// // RBAC: Make sure the RoleBinding exists
	// err = r.CreateOrUpdateRoleBindingStatus(builder)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// // Exit Reconciliation Loop
	// logger.Info("Ending reconcile loop")

	// FINAL: Set Status.Ready state
	err = r.SetReadyStatus(ctx, resource)
	if err != nil {
		return ctrl.Result{}, err
	}

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

//
// func (r *ExecAccessRequestReconciler) CreateOrUpdateRoleStatus(builder *builders.AccessBuilder) error {
// 	logger := r.GetLogger(builder.Ctx)
//
// 	// Get a representation of the role that we want.
// 	role, _ := builder.GenerateAccessRole()
// 	emptyRole := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: role.Name, Namespace: role.Namespace}}
//
// 	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
// 	op, err := controllerutil.CreateOrUpdate(builder.Ctx, builder.Client, emptyRole, func() error {
// 		emptyRole.ObjectMeta = role.ObjectMeta
// 		emptyRole.Rules = role.Rules
// 		emptyRole.OwnerReferences = role.OwnerReferences
// 		return nil
// 	})
//
// 	// If there was an error, log it, and update the conditions appropriately
// 	if err != nil {
// 		logger.Error(err, fmt.Sprintf("Failure reconciling role %s", role.Name), "operation", op)
// 		if err := r.UpdateCondition(
// 			builder.Ctx, builder.Request,
// 			ConditionRoleCreated,
// 			metav1.ConditionFalse,
// 			string(metav1.StatusFailure),
// 			fmt.Sprintf("ERROR: %s", err)); err != nil {
// 			return err
// 		}
// 	}
//
// 	// Success, update the object condition
// 	if err := r.UpdateCondition(
// 		builder.Ctx, builder.Request,
// 		ConditionRoleCreated,
// 		metav1.ConditionTrue,
// 		string(metav1.StatusSuccess),
// 		"Role successfully reconciled"); err != nil {
// 		return err
// 	}
//
// 	// Update the request's RoleName field
// 	builder.Request.Status.RoleName = role.Name
// 	if err = r.UpdateStatus(builder.Ctx, builder.Request); err != nil {
// 		return err
// 	}
//
// 	logger.Info(fmt.Sprintf("Role %s successfully reconciled", role.Name), "operation", op)
// 	return nil
// }
//
// func (r *ExecAccessRequestReconciler) CreateOrUpdateRoleBindingStatus(builder *builders.AccessBuilder) error {
// 	logger := r.GetLogger(builder.Ctx)
//
// 	// Get a representation of the role that we want.
// 	rb, _ := builder.GenerateAccessRoleBinding()
// 	emptyRb := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: rb.Name, Namespace: rb.Namespace}}
//
// 	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
// 	op, err := controllerutil.CreateOrUpdate(builder.Ctx, builder.Client, emptyRb, func() error {
// 		emptyRb.ObjectMeta = rb.ObjectMeta
// 		emptyRb.RoleRef = rb.RoleRef
// 		emptyRb.Subjects = rb.Subjects
// 		emptyRb.OwnerReferences = rb.OwnerReferences
// 		return nil
// 	})
//
// 	// If there was an error, log it, and update the conditions appropriately
// 	if err != nil {
// 		logger.Error(err, fmt.Sprintf("Failure reconciling rolebinding %s", rb.Name), "operation", op)
// 		if err := r.UpdateCondition(
// 			builder.Ctx, builder.Request,
// 			ConditionRoleBindingCreated,
// 			metav1.ConditionFalse,
// 			string(metav1.StatusFailure),
// 			fmt.Sprintf("ERROR: %s", err)); err != nil {
// 			return err
// 		}
// 	}
//
// 	// Success, update the object condition
// 	if err := r.UpdateCondition(
// 		builder.Ctx, builder.Request,
// 		ConditionRoleBindingCreated,
// 		metav1.ConditionTrue,
// 		string(metav1.StatusSuccess),
// 		"RoleBinding successfully reconciled"); err != nil {
// 		return err
// 	}
//
// 	// Update the request's RoleName field
// 	builder.Request.Status.RoleBindingName = rb.Name
// 	if err = r.UpdateStatus(builder.Ctx, builder.Request); err != nil {
// 		return err
// 	}
//
// 	logger.Info(fmt.Sprintf("RoleBinding %s successfully reconciled", rb.Name), "operation", op)
// 	return nil
// }
//
// func (r *ExecAccessRequestReconciler) VerifyTargetPodExists(ctx context.Context, req *api.ExecAccessRequest, podName string) error {
// 	logger := r.GetLogger(ctx)
// 	logger.Info(fmt.Sprintf("Verifying that Pod %s still exists...", podName))
//
// 	// Search for the Pod
// 	pod := &v1.Pod{}
// 	err := r.Get(ctx, types.NamespacedName{
// 		Name:      podName,
// 		Namespace: req.GetNamespace(),
// 	}, pod)
//
// 	// On any failure, update the pod status with the failure...
// 	if err != nil {
// 		logger.Info(fmt.Sprintf("Pod %s is missing. Updating status.", podName))
// 		return r.UpdateCondition(
// 			ctx, req,
// 			ConditionTargetPodExists,
// 			metav1.ConditionUnknown,
// 			string(metav1.StatusReasonNotFound),
// 			fmt.Sprintf("ERROR: %s", err),
// 		)
// 	}
// 	return r.UpdateCondition(
// 		ctx, req,
// 		ConditionTargetPodExists,
// 		metav1.ConditionTrue,
// 		string(metav1.StatusSuccess),
// 		fmt.Sprintf("Found Pod (UID: %s)", pod.UID),
// 	)
// }

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
