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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers/builders"
)

// ExecAccessTemplateReconciler reconciles a ExecAccessTemplate object
type ExecAccessTemplateReconciler struct {
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
	log := log.FromContext(ctx)
	log.Info("Starting reconcile loop")

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := r.GetLogger(ctx)
	logger.Info("Starting reconcile loop")

	// Get the ExecAccessTemplate resource if it exists. If not, we bail out quietly.
	//
	// TODO: If this resource is deleted, then we need to find all AccessRequests pointing to it,
	// and delete them as well.
	tmpl, err := getExecAccessTemplate(r.Client, ctx, req.Name, req.Namespace)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to find ExecAccessTemplate %s, perhaps deleted.", req))
		return ctrl.Result{}, nil
	}

	// Create an ExecAccessBuilder resource for this particular template, which we'll use to then verify the resource.
	builder := &builders.ExecAccessBuilder{
		Client:   r.Client,
		Ctx:      ctx,
		Scheme:   r.Scheme,
		Template: tmpl,
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

	// TODO:
	// VERIFICATION: Ensure that the allowedGroups match valid group name strings

	return ctrl.Result{}, nil
}

func (r *ExecAccessTemplateReconciler) VerifyMiscSettings(builder *builders.ExecAccessBuilder) error {
	// Verify that MaxDuration is greater than DesiredDuration.
	defaultDuration, err := builder.Template.GetDefaultDuration()
	if err != nil {
		return r.UpdateCondition(
			builder.Ctx, builder.Template, ConditionDurationsValid, metav1.ConditionFalse,
			string(metav1.StatusReasonNotAcceptable), fmt.Sprintf("Error on spec.defaultDuration: %s", err))
	}
	maxDuration, err := builder.Template.GetMaxDuration()
	if err != nil {
		return r.UpdateCondition(
			builder.Ctx, builder.Template, ConditionDurationsValid, metav1.ConditionFalse,
			string(metav1.StatusReasonNotAcceptable), fmt.Sprintf("Error on spec.maxDuration: %s", err))
	}
	if defaultDuration > maxDuration {
		return r.UpdateCondition(
			builder.Ctx, builder.Template, ConditionDurationsValid, metav1.ConditionFalse,
			string(metav1.StatusReasonNotAcceptable),
			"Error: spec.defaultDuration can not be greater than spec.maxDuration")
	} else {
		return r.UpdateCondition(
			builder.Ctx, builder.Template, ConditionDurationsValid, metav1.ConditionTrue,
			string(metav1.StatusSuccess),
			"spec.defaultDuration and spec.maxDuration valid")
	}
}

func (r *ExecAccessTemplateReconciler) VerifyTargetRef(builder *builders.ExecAccessBuilder) error {
	targetRef := builder.Template.Spec.TargetRef
	var err error
	if targetRef.Kind == api.DeploymentController {
		_, err = builder.GetDeployment()
	} else if targetRef.Kind == api.DaemonSetController {
		_, err = builder.GetDaemonSet()
	} else if targetRef.Kind == api.StatefulSetController {
		_, err = builder.GetStatefulSet()
	}

	if err != nil {
		return r.UpdateCondition(
			builder.Ctx, builder.Template, ConditionTargetRefExists, metav1.ConditionFalse,
			string(metav1.StatusReasonNotFound), fmt.Sprintf("Error: %s", err))
	}

	return r.UpdateCondition(
		builder.Ctx, builder.Template, ConditionTargetRefExists, metav1.ConditionTrue,
		string(metav1.StatusSuccess), "Success")
}

func (r *ExecAccessTemplateReconciler) UpdateCondition(
	ctx context.Context,
	req *api.ExecAccessTemplate,
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
func (r *ExecAccessTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ExecAccessTemplate{}).
		Complete(r)
}

// GetResource returns back an ExecAccessTemplate resource matching the request supplied to the reconciler loop, or
// returns back an error.
func getExecAccessTemplate(cl client.Client, ctx context.Context, name string, namespace string) (*api.ExecAccessTemplate, error) {
	tmpl := &api.ExecAccessTemplate{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
