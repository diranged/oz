package controllers

import (
	"fmt"

	"github.com/diranged/oz/controllers/builders"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OzTemplateReconciler provides a base reconciler with common functions for handling our Template CRDs
// (ExecAccessTemplate, AccessTemplate, etc)
type OzTemplateReconciler struct {
	OzReconciler
}

// VerifyTargetRef ensures that the Spec.targetRef points to a valid and understood controller that we
// can build our templates off of. Any failure results in the resource ConditionTargetRefExists condition
// being set to False.
//
// Returns:
//   - An "error" only if the UpdateCondition function fails
func (r *OzTemplateReconciler) VerifyTargetRef(builder builders.Builder) error {
	var err error
	ctx := builder.GetCtx()
	tmpl := builder.GetTemplate()

	logger := log.FromContext(builder.GetCtx())
	logger.Info("Beginning TargetRef Verification")

	targetRef, err := builder.GetTargetRefResource()
	if err != nil {
		return r.updateCondition(
			ctx, tmpl, conditionTargetRefExists, metav1.ConditionFalse,
			string(metav1.StatusReasonNotFound), fmt.Sprintf("Error: %s", err))
	}

	logger.Info(fmt.Sprintf("Returning %s", targetRef.GetObjectKind().GroupVersionKind().Kind))
	return r.updateCondition(
		ctx, tmpl, conditionTargetRefExists, metav1.ConditionTrue,
		string(metav1.StatusSuccess), "Success")
}

// VerifyMiscSettings walks through the common required settings for any OzTemplateResource. For
// each setting we will update an appropriate Condition within the resource.
//
// Returns:
//   - An "error" only if the UpdateCondition function fails
func (r *OzTemplateReconciler) VerifyMiscSettings(builder builders.Builder) error {
	ctx := builder.GetCtx()
	tmpl := builder.GetTemplate()

	// Verify that MaxDuration is greater than DesiredDuration.
	defaultDuration, err := tmpl.GetAccessConfig().GetDefaultDuration()
	if err != nil {
		return r.updateCondition(
			ctx, tmpl, conditionDurationsValid, metav1.ConditionFalse,
			string(metav1.StatusReasonNotAcceptable), fmt.Sprintf("Error on spec.defaultDuration: %s", err))
	}
	maxDuration, err := tmpl.GetAccessConfig().GetMaxDuration()
	if err != nil {
		return r.updateCondition(
			ctx, tmpl, conditionDurationsValid, metav1.ConditionFalse,
			string(metav1.StatusReasonNotAcceptable), fmt.Sprintf("Error on spec.maxDuration: %s", err))
	}
	if defaultDuration > maxDuration {
		return r.updateCondition(
			ctx, tmpl, conditionDurationsValid, metav1.ConditionFalse,
			string(metav1.StatusReasonNotAcceptable),
			"Error: spec.defaultDuration can not be greater than spec.maxDuration")
	}
	return r.updateCondition(
		ctx, tmpl, conditionDurationsValid, metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		"spec.defaultDuration and spec.maxDuration valid")
}
