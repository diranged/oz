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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		return tmpl, r.UpdateCondition(
			ctx, req, ConditionTargetTemplateExists, metav1.ConditionTrue, string(metav1.StatusSuccess),
			"Found Target Template")
	}
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
			logger.Info("New Pod Name", "PodName", builder.Request.Status.PodName)
			return builder.Request.Status.PodName, err
		}
		return builder.Request.Status.PodName, nil
	}
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
