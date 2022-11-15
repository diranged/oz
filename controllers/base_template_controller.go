package controllers

import (
	"context"
	"fmt"

	api "github.com/diranged/oz/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// OzTemplateReconciler provides a base reconciler with common functions for handling our Template CRDs
// (ExecAccessTemplate, AccessTemplate, etc)
type OzTemplateReconciler struct {
	OzReconciler
}

func (r *OzTemplateReconciler) VerifyTargetTemplate(ctx context.Context, req *api.ExecAccessRequest) (*api.ExecAccessTemplate, error) {
	logger := r.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Verifying that Target Template %s still exists...", req.Spec.TemplateName))
	if tmpl, err := api.GetExecAccessTemplate(r.Client, ctx, req.Spec.TemplateName, req.Namespace); err != nil {
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
