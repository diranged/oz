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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers/builders"
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
	_ = log.FromContext(ctx)

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := r.GetLogger(ctx)
	logger.Info("Starting reconcile loop")

	// TODO: If this resource is deleted, then we need to find all AccessRequests pointing to it,
	// and delete them as well.
	request, err := getExecAccessRequest(r.Client, ctx, req.Name, req.Namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to find ExecAccessRequest %s, perhaps deleted.", req))
		return ctrl.Result{}, nil
	}

	// VERIFICATION: Make sure the Target TemplateName field points to a valid Template
	tmpl, err := r.VerifyTargetTemplate(ctx, request)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create an ExecAccessBuilder resource for this particular template, which we'll use to then verify the resource.
	builder := &builders.ExecAccessBuilder{
		Client:   r.Client,
		Ctx:      ctx,
		Scheme:   r.Scheme,
		Request:  request,
		Template: tmpl,
	}

	// Get or Set the Target Pod Name for the access request. If the Status.TargetPod field is already set, this
	// will simply return that value.
	_, err = r.GetOrSetPodNameStatus(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// VERIFICATION: Make sure the Target Pod still exists - that it hasn't gone away at some point.
	r.VerifyTargetPodExists(ctx, request, request.Status.PodName)

	// VERIFICATION: Verifies the requested duration
	err = r.VerifyDuration(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// RBAC: Make sure the Role exists
	err = r.CreateOrUpdateRoleStatus(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// RBAC: Make sure the RoleBinding exists
	err = r.CreateOrUpdateRoleBindingStatus(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// FINAL: Handle whether or not the access is expired at this point! If so, delete it.
	err = r.HandleAccessExpired(builder)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Exit Reconciliation Loop
	logger.Info("Ending reconcile loop")
	return ctrl.Result{}, nil
}

func (r *ExecAccessRequestReconciler) VerifyTargetTemplate(ctx context.Context, req *api.ExecAccessRequest) (*api.ExecAccessTemplate, error) {
	logger := r.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Verifying that Target Template %s still exists...", req.Spec.TemplateName))
	if tmpl, err := getExecAccessTemplate(r.Client, ctx, req.Spec.TemplateName, req.Namespace); err != nil {
		return nil, r.UpdateCondition(
			ctx, req, ConditionTargetTemplateExists, metav1.ConditionFalse,
			string(metav1.StatusReasonNotFound), fmt.Sprintf("Error: %s", err))
	} else {
		// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
		//
		// Ensure that if the TargetTemplate is ever deleted, that all of the AccessRequests are
		// also deleted, which will cascade down and delete any roles/bindings/etc.
		if err := ctrl.SetControllerReference(tmpl, req, r.Scheme); err != nil {
			return nil, err
		}

		return tmpl, r.UpdateCondition(
			ctx, req, ConditionTargetTemplateExists, metav1.ConditionTrue, string(metav1.StatusSuccess),
			"Found Target Template")
	}
}

func (r *ExecAccessRequestReconciler) VerifyDuration(builder *builders.ExecAccessBuilder) error {
	var err error
	logger := r.GetLogger(builder.Ctx)
	logger.Info("Beginning access request duration verification")

	// Step one - verify the inputs themselves. If the user supplied invalid inputs, or the template has any
	// invalid inputs, we bail out and update the conditions as such. This is to prevent escalated privilegess
	// from lasting indefinitely.
	var requestedDuration time.Duration
	if builder.Request.Spec.Duration != "" {
		requestedDuration, err = builder.Request.GetDuration()
		if err != nil {
			r.UpdateCondition(builder.Ctx, builder.Request, ConditionDurationsValid,
				metav1.ConditionFalse, string(metav1.StatusReasonBadRequest), fmt.Sprintf("spec.duration error: %s", err))
			return err
		}
	}
	templateDefaultDuration, err := builder.Template.GetDefaultDuration()
	if err != nil {
		r.UpdateCondition(builder.Ctx, builder.Request, ConditionDurationsValid,
			metav1.ConditionFalse, string(metav1.StatusReasonBadRequest), fmt.Sprintf("Template Error, spec.defaultDuration error: %s", err))
		return err
	}

	templateMaxDuration, err := builder.Template.GetMaxDuration()
	if err != nil {
		r.UpdateCondition(builder.Ctx, builder.Request, ConditionDurationsValid,
			metav1.ConditionFalse, string(metav1.StatusReasonBadRequest), fmt.Sprintf("Template Error, spec.maxDuration error: %s", err))
		return err
	}

	// Now determine which duration is the one we'll use
	var accessDuration time.Duration
	var reasonStr string

	if requestedDuration == 0 {
		// If no requested duration supplied, then default to the template's default duration
		reasonStr = fmt.Sprintf("Access request duration defaulting to template duration time (%s)", templateDefaultDuration.String())
		accessDuration = templateDefaultDuration
	} else if requestedDuration <= templateMaxDuration {
		// If the requested duration is too long, use the template max
		reasonStr = fmt.Sprintf("Access requested custom duration (%s)", requestedDuration.String())
		accessDuration = requestedDuration
	} else {
		// Finally, if it's valid, use the supplied duration
		reasonStr = fmt.Sprintf("Access requested duration (%s) larger than template maximum duration (%s)", requestedDuration.String(), templateMaxDuration.String())
		accessDuration = templateMaxDuration
	}

	// Log out the decision, and update the condition
	logger.Info(reasonStr)

	// TESTING: Trying to make sure the below updatecondition doesn't fail
	r.Refetch(builder.Ctx, builder.Request)

	err = r.UpdateCondition(builder.Ctx, builder.Request, ConditionDurationsValid,
		metav1.ConditionTrue, string(metav1.StatusSuccess), reasonStr)
	if err != nil {
		return err
	}

	// Determine how long the AccessRequest has been around, and compare that to the accessDuration.
	now := time.Now()
	creation := builder.Request.CreationTimestamp.Time
	accessUptime := now.Sub(creation)

	// If the accessUptime is greater than the accessDuration, kill it.
	if accessUptime > accessDuration {
		return r.UpdateCondition(builder.Ctx, builder.Request, ConditionAccessStillValid,
			metav1.ConditionFalse, string(metav1.StatusReasonTimeout), "Access expired")
	}

	return r.UpdateCondition(builder.Ctx, builder.Request, ConditionAccessStillValid,
		metav1.ConditionTrue, string(metav1.StatusReasonTimeout), "Access still valid")
}

func (r *ExecAccessRequestReconciler) HandleAccessExpired(builder *builders.ExecAccessBuilder) error {
	logger := r.GetLogger(builder.Ctx)
	logger.Info("Checking if access has expired or not...")
	cond := meta.FindStatusCondition(builder.Request.Status.Conditions, string(ConditionAccessStillValid))
	if cond == nil {
		logger.Info(fmt.Sprintf("Missing Condition %s, skipping deletion", ConditionAccessStillValid))
		return nil
	}

	if cond.Status == metav1.ConditionFalse {
		logger.Info(fmt.Sprintf("Found Condition %s in state %s, terminating rqeuest", ConditionAccessStillValid, cond.Status))
		return r.DeleteResource(builder)
	}

	logger.Info(fmt.Sprintf("Found Condition %s in state %s, leaving alone", ConditionAccessStillValid, cond.Status))

	return nil
}

func (r *ExecAccessRequestReconciler) DeleteResource(builder *builders.ExecAccessBuilder) error {
	return r.Delete(builder.Ctx, builder.Request)
}

func (r *ExecAccessRequestReconciler) GetOrSetPodNameStatus(builder *builders.ExecAccessBuilder) (string, error) {
	logger := r.GetLogger(builder.Ctx)

	// If this resource already has a status.podName field set, then we respect that no matter what.
	// We never mutate the pod that this access request was originally created for. Otherwise, pick
	// a Pod and populate that status field.
	if builder.Request.Status.PodName != "" {
		logger.Info(fmt.Sprintf("Pod already assigned - %s", builder.Request.Status.PodName))
		return builder.Request.Status.PodName, nil
	}

	if podName, err := builder.GetTargetPodName(); err != nil {
		return "", err
	} else {
		if podName != "" && podName != builder.Request.Status.PodName {
			builder.Request.Status.PodName = podName
			err = r.UpdateStatus(builder.Ctx, builder.Request)
			logger.Info(fmt.Sprintf("Target Pod Name %s", builder.Request.Status.PodName))
			return builder.Request.Status.PodName, err
		}
		return builder.Request.Status.PodName, nil
	}
}

func (r *ExecAccessRequestReconciler) CreateOrUpdateRoleStatus(builder *builders.ExecAccessBuilder) error {
	logger := r.GetLogger(builder.Ctx)

	// Get a representation of the role that we want.
	role, _ := builder.GenerateAccessRole()
	emptyRole := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: role.Name, Namespace: role.Namespace}}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	op, err := controllerutil.CreateOrUpdate(builder.Ctx, builder.Client, emptyRole, func() error {
		emptyRole.ObjectMeta = role.ObjectMeta
		emptyRole.Rules = role.Rules
		emptyRole.OwnerReferences = role.OwnerReferences
		return nil
	})

	// If there was an error, log it, and update the conditions appropriately
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failure reconciling role %s", role.Name), "operation", op)
		if err := r.UpdateCondition(
			builder.Ctx, builder.Request,
			ConditionRoleCreated,
			metav1.ConditionFalse,
			string(metav1.StatusFailure),
			fmt.Sprintf("ERROR: %s", err)); err != nil {
			return err
		}
	}

	// Success, update the object condition
	if err := r.UpdateCondition(
		builder.Ctx, builder.Request,
		ConditionRoleCreated,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		"Role successfully reconciled"); err != nil {
		return err
	}

	// Update the request's RoleName field
	builder.Request.Status.RoleName = role.Name
	if err = r.UpdateStatus(builder.Ctx, builder.Request); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Role %s successfully reconciled", role.Name), "operation", op)
	return nil
}

func (r *ExecAccessRequestReconciler) CreateOrUpdateRoleBindingStatus(builder *builders.ExecAccessBuilder) error {
	logger := r.GetLogger(builder.Ctx)

	// Get a representation of the role that we want.
	rb, _ := builder.GenerateAccessRoleBinding()
	emptyRb := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: rb.Name, Namespace: rb.Namespace}}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	op, err := controllerutil.CreateOrUpdate(builder.Ctx, builder.Client, emptyRb, func() error {
		emptyRb.ObjectMeta = rb.ObjectMeta
		emptyRb.RoleRef = rb.RoleRef
		emptyRb.Subjects = rb.Subjects
		emptyRb.OwnerReferences = rb.OwnerReferences
		return nil
	})

	// If there was an error, log it, and update the conditions appropriately
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failure reconciling rolebinding %s", rb.Name), "operation", op)
		if err := r.UpdateCondition(
			builder.Ctx, builder.Request,
			ConditionRoleBindingCreated,
			metav1.ConditionFalse,
			string(metav1.StatusFailure),
			fmt.Sprintf("ERROR: %s", err)); err != nil {
			return err
		}
	}

	// Success, update the object condition
	if err := r.UpdateCondition(
		builder.Ctx, builder.Request,
		ConditionRoleBindingCreated,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		"RoleBinding successfully reconciled"); err != nil {
		return err
	}

	// Update the request's RoleName field
	builder.Request.Status.RoleBindingName = rb.Name
	if err = r.UpdateStatus(builder.Ctx, builder.Request); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("RoleBinding %s successfully reconciled", rb.Name), "operation", op)
	return nil
}

func (r *ExecAccessRequestReconciler) VerifyTargetPodExists(ctx context.Context, req *api.ExecAccessRequest, podName string) error {
	logger := r.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Verifying that Pod %s still exists...", podName))

	// Search for the Pod
	pod := &v1.Pod{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: req.GetNamespace(),
	}, pod)

	// On any failure, update the pod status with the failure...
	if err != nil {
		logger.Info(fmt.Sprintf("Pod %s is missing. Updating status.", podName))
		return r.UpdateCondition(
			ctx, req,
			ConditionTargetPodExists,
			metav1.ConditionUnknown,
			string(metav1.StatusReasonNotFound),
			fmt.Sprintf("ERROR: %s", err),
		)
	}
	return r.UpdateCondition(
		ctx, req,
		ConditionTargetPodExists,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		fmt.Sprintf("Found Pod (UID: %s)", pod.UID),
	)
}

func (r *ExecAccessRequestReconciler) UpdateCondition(
	ctx context.Context,
	req *api.ExecAccessRequest,
	conditionType RequestConditionTypes,
	conditionStatus metav1.ConditionStatus,
	reason string,
	message string,
) error {
	meta.SetStatusCondition(&req.Status.Conditions, metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		ObservedGeneration: req.GetGeneration(),
		LastTransitionTime: metav1.Time{},
		Reason:             reason,
		Message:            message,
	})
	err := r.UpdateStatus(ctx, req)
	return err

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

// GetResource returns back an ExecAccessRequest resource matching the request supplied to the reconciler loop, or
// returns back an error.
func getExecAccessRequest(cl client.Client, ctx context.Context, name string, namespace string) (*api.ExecAccessRequest, error) {
	tmpl := &api.ExecAccessRequest{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}
