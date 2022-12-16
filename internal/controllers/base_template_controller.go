package controllers

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/controllers/internal/status"
	"github.com/diranged/oz/internal/legacybuilder"
)

// BaseTemplateReconciler provides a base reconciler with common functions for handling our Template CRDs
// (ExecAccessTemplate, AccessTemplate, etc)
type BaseTemplateReconciler struct {
	BaseReconciler
}

// VerifyTargetRef ensures that the Spec.targetRef points to a valid and understood controller that we
// can build our templates off of. Any failure results in the resource ConditionTargetRefExists condition
// being set to False.
//
// Returns:
//   - An "error" only if the UpdateCondition function fails
func (r *BaseTemplateReconciler) VerifyTargetRef(builder legacybuilder.IBuilder) error {
	var err error
	ctx := builder.GetCtx()
	tmpl := builder.GetTemplate()

	logger := log.FromContext(builder.GetCtx())
	logger.Info("Beginning TargetRef Verification")

	targetRef, err := builder.GetTargetRefResource()
	if err != nil {
		return status.SetTargetRefNotExists(ctx, r, tmpl, err)
	}

	logger.Info(fmt.Sprintf("Returning %s", targetRef.GetObjectKind().GroupVersionKind().Kind))
	return status.SetTargetRefExists(ctx, r, tmpl, "Success")
}

// VerifyMiscSettings walks through the common required settings for any OzTemplateResource. For
// each setting we will update an appropriate Condition within the resource.
//
// Returns:
//   - An "error" only if the UpdateCondition function fails
func (r *BaseTemplateReconciler) VerifyMiscSettings(builder legacybuilder.IBuilder) error {
	ctx := builder.GetCtx()
	tmpl := builder.GetTemplate()

	// Verify that MaxDuration is greater than DesiredDuration.
	defaultDuration, err := tmpl.GetAccessConfig().GetDefaultDuration()
	if err != nil {
		return status.SetTemplateDurationsNotValid(ctx, r, tmpl,
			fmt.Sprintf("Error on spec.defaultDuration: %s", err),
		)
	}
	maxDuration, err := tmpl.GetAccessConfig().GetMaxDuration()
	if err != nil {
		return status.SetTemplateDurationsNotValid(ctx, r, tmpl,
			fmt.Sprintf("Error on spec.maxDuration: %s", err),
		)
	}
	if defaultDuration > maxDuration {
		return status.SetTemplateDurationsNotValid(ctx, r, tmpl,
			"Error: spec.defaultDuration can not be greater than spec.maxDuration")
	}
	return status.SetTemplateDurationsValid(ctx, r, tmpl,
		"spec.defaultDuration and spec.maxDuration valid",
	)
}
