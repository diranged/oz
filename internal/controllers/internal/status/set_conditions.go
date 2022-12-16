package status

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

/*
IRequestResource Condition Setters
*/

// SetTargetTemplateExists sets the ConditionTargetTemplateExists condition to True
func SetTargetTemplateExists(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionTargetTemplateExists,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		"Found Target Template",
	)
}

// SetTargetTemplateNotExists sets the ConditionTargetTemplateExists condition to False
func SetTargetTemplateNotExists(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	err error,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionTargetTemplateExists,
		metav1.ConditionFalse,
		string(metav1.StatusReasonNotFound),
		fmt.Sprintf("Error: %s", err),
	)
}

// SetRequestDurationsNotValid updates the ConditionRequestDurationsValid
// condition on a Request resource to a failure.
func SetRequestDurationsNotValid(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	reason string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionRequestDurationsValid,
		metav1.ConditionFalse,
		string(metav1.StatusReasonBadRequest),
		reason,
	)
}

// SetRequestDurationsValid updates the ConditionRequestDurationsValid
// condition on a Request resource to a success.
func SetRequestDurationsValid(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	reason string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionRequestDurationsValid,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		reason,
	)
}

// SetAccessNotValid updates the ConditionAccessStillValid condition to False.
func SetAccessNotValid(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionAccessStillValid,
		metav1.ConditionFalse,
		string(metav1.StatusReasonTimeout),
		"Access expired",
	)
}

// SetAccessStillValid updates the ConditionAccessStillValid condition to True.
func SetAccessStillValid(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionAccessStillValid,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		"Access still valid",
	)
}

// SetAccessResourcesNotCreated updates the ConditionAccessResourcesCreated condition to False.
func SetAccessResourcesNotCreated(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	err error,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionAccessResourcesCreated,
		metav1.ConditionFalse,
		string(metav1.StatusFailure),
		fmt.Sprintf("ERROR: %s", err),
	)
}

// SetAccessResourcesCreated updates the ConditionAccessResourcesCreated condition to True.
func SetAccessResourcesCreated(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	message string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionAccessResourcesCreated,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		message)
}

// SetAccessResourcesNotReady updates the ConditionAccessResourcesReady condition to False.
func SetAccessResourcesNotReady(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	err error,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionAccessResourcesReady,
		metav1.ConditionFalse,
		"NotYetReady",
		fmt.Sprintf("%s", err),
	)
}

// SetAccessResourcesReady updates the ConditionAccessResourcesReady condition to True.
func SetAccessResourcesReady(
	ctx context.Context,
	rec hasStatusReconciler,
	req v1alpha1.IRequestResource,
	message string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		req,
		v1alpha1.ConditionAccessResourcesReady,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		message)
}

/*
ITemplateResource Condition Setters
*/

// SetTargetRefExists updates the ConditionTargetRefExists condition on a
// Template resource to success.
func SetTargetRefExists(
	ctx context.Context,
	rec hasStatusReconciler,
	tmpl v1alpha1.ITemplateResource,
	message string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		tmpl,
		v1alpha1.ConditionTargetRefExists,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		message,
	)
}

// SetTargetRefNotExists updates the ConditionTargetRefExists condition on a
// Template resource to a failure based on the Error supplied.
func SetTargetRefNotExists(
	ctx context.Context,
	rec hasStatusReconciler,
	tmpl v1alpha1.ITemplateResource,
	err error,
) error {
	return UpdateCondition(
		ctx,
		rec,
		tmpl,
		v1alpha1.ConditionTargetRefExists,
		metav1.ConditionFalse,
		string(metav1.StatusReasonNotFound),
		fmt.Sprintf("Error: %s", err),
	)
}

// SetTemplateDurationsNotValid updates the ConditionTemplateDurationsValid
// condition on a Template resource to a failure.
func SetTemplateDurationsNotValid(
	ctx context.Context,
	rec hasStatusReconciler,
	tmpl v1alpha1.ITemplateResource,
	reason string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		tmpl,
		v1alpha1.ConditionTemplateDurationsValid,
		metav1.ConditionFalse,
		string(metav1.StatusReasonNotAcceptable),
		reason,
	)
}

// SetTemplateDurationsValid updates the ConditionTemplateDurationsValid
// condition on a Template resource to a success.
func SetTemplateDurationsValid(
	ctx context.Context,
	rec hasStatusReconciler,
	tmpl v1alpha1.ITemplateResource,
	reason string,
) error {
	return UpdateCondition(
		ctx,
		rec,
		tmpl,
		v1alpha1.ConditionTemplateDurationsValid,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		reason,
	)
}
