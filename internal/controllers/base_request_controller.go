package controllers

import (
	"fmt"
	"time"

	"github.com/diranged/oz/controllers/builders"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BaseRequestReconciler provides a base reconciler with common functions for handling our Template CRDs
// (ExecAccessTemplate, AccessTemplate, etc)
type BaseRequestReconciler struct {
	BaseReconciler
}

// verifyDuration checks a few components of whether or not the AccessRequest is still valid:
//
//   - Was the (optional) supplied "spec.duration" valid?
//   - Is the target tempate "spec.defaultDuration"  valid?
//   - Is the target template "spec.maxDuration" valid?
//   - Did the user supply their own "spec.duration"?
//     yes? Is it lower than the target template "spec.maxDuration"?
//     no? Use the target template "spec.defaultDuration"
//   - Is the access request duration less than its current age?
//     yes? approve
//     no? mark the resource for deletion
func (r *BaseRequestReconciler) verifyDuration(builder builders.IBuilder) error {
	var err error
	logger := r.getLogger(builder.GetCtx())

	logger.Info("Beginning access request duration verification")

	// Step one - verify the inputs themselves. If the user supplied invalid inputs, or the template has any
	// invalid inputs, we bail out and update the conditions as such. This is to prevent escalated privilegess
	// from lasting indefinitely.
	var requestedDuration time.Duration
	if requestedDuration, err = builder.GetRequest().GetDuration(); err != nil {
		// TODO: check err return from updateCondition
		_ = r.updateCondition(
			builder.GetCtx(),
			builder.GetRequest(),
			ConditionDurationsValid,
			metav1.ConditionFalse,
			string(metav1.StatusReasonBadRequest),
			fmt.Sprintf("spec.duration error: %s", err),
		)
		return err
	}
	templateDefaultDuration, err := builder.GetTemplate().GetAccessConfig().GetDefaultDuration()
	if err != nil {
		// TODO: check err return from updateCondition
		_ = r.updateCondition(
			builder.GetCtx(),
			builder.GetRequest(),
			ConditionDurationsValid,
			metav1.ConditionFalse,
			string(metav1.StatusReasonBadRequest),
			fmt.Sprintf("Template Error, spec.defaultDuration error: %s", err),
		)
		return err
	}

	templateMaxDuration, err := builder.GetTemplate().GetAccessConfig().GetMaxDuration()
	if err != nil {
		// TODO: check err return from updateCondition
		_ = r.updateCondition(
			builder.GetCtx(),
			builder.GetRequest(),
			ConditionDurationsValid,
			metav1.ConditionFalse,
			string(metav1.StatusReasonBadRequest),
			fmt.Sprintf("Template Error, spec.maxDuration error: %s", err),
		)
		return err
	}

	// Now determine which duration is the one we'll use
	var accessDuration time.Duration
	var reasonStr string

	if requestedDuration == 0 {
		// If no requested duration supplied, then default to the template's default duration
		reasonStr = fmt.Sprintf(
			"Access request duration defaulting to template duration time (%s)",
			templateDefaultDuration.String(),
		)
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

	err = r.updateCondition(builder.GetCtx(), builder.GetRequest(), ConditionDurationsValid,
		metav1.ConditionTrue, string(metav1.StatusSuccess), reasonStr)
	if err != nil {
		return err
	}

	// If the accessUptime is greater than the accessDuration, kill it.
	if builder.GetRequest().GetUptime() > accessDuration {
		return r.updateCondition(builder.GetCtx(), builder.GetRequest(), ConditionAccessStillValid,
			metav1.ConditionFalse, string(metav1.StatusReasonTimeout), "Access expired")
	}

	// Update the resource, and let the user know how much time is remaining
	return r.updateCondition(builder.GetCtx(), builder.GetRequest(), ConditionAccessStillValid,
		metav1.ConditionTrue, string(metav1.StatusSuccess),
		"Access still valid")
}

// isAccessExpired checks the AccessRequest status for the ConditionAccessStillValid condition. If it is no longer
// a valid request, then the resource is immediately deleted.
//
// Returns:
//
//	true: if the resource is expired, AND has now been deleted
//	false: if the resource is still valid
//	error: any error during the checks
func (r *BaseRequestReconciler) isAccessExpired(builder builders.IBuilder) (bool, error) {
	logger := r.getLogger(builder.GetCtx())
	logger.Info("Checking if access has expired or not...")
	cond := meta.FindStatusCondition(
		*builder.GetRequest().GetStatus().GetConditions(),
		string(ConditionAccessStillValid),
	)
	if cond == nil {
		logger.Info(
			fmt.Sprintf("Missing Condition %s, skipping deletion", ConditionAccessStillValid),
		)
		return false, nil
	}

	if cond.Status == metav1.ConditionFalse {
		logger.Info(
			fmt.Sprintf(
				"Found Condition %s in state %s, terminating rqeuest",
				ConditionAccessStillValid,
				cond.Status,
			),
		)
		return true, r.DeleteResource(builder)
	}

	logger.Info(
		fmt.Sprintf(
			"Found Condition %s in state %s, leaving alone",
			ConditionAccessStillValid,
			cond.Status,
		),
	)
	return false, nil
}

// verifyAccessResourcesBuilt calls out to the Builder interface's GenerateAccessResources() method to build out
// all of the resources that are required for thie particular access request. The Status.Conditions field is
// then updated with the ConditionAccessResourcesCreated condition appropriately.
func (r *BaseRequestReconciler) verifyAccessResourcesBuilt(
	builder builders.IBuilder,
) error {
	logger := log.FromContext(builder.GetCtx())
	logger.Info("Verifying that access resources are built")

	statusString, err := builder.GenerateAccessResources()
	if err != nil {
		// TODO: check err return from updateCondition
		_ = r.updateCondition(
			builder.GetCtx(), builder.GetRequest(),
			ConditionAccessResourcesCreated,
			metav1.ConditionFalse,
			string(metav1.StatusFailure),
			fmt.Sprintf("ERROR: %s", err))
		return err
	}
	return r.updateCondition(
		builder.GetCtx(), builder.GetRequest(),
		ConditionAccessResourcesCreated,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		statusString)
}

// verifyAccessResourcesReady is a followup to the verifyAccessResources()
// function - where we make sure that the .Status.PodName resource has come all
// the way up and reached the "Running" phase.
func (r *BaseRequestReconciler) verifyAccessResourcesReady(
	builder builders.IPodAccessBuilder,
) error {
	logger := log.FromContext(builder.GetCtx())
	logger.Info("Verifying that access resources are ready")

	statusString, err := builder.VerifyAccessResources()
	if err != nil {
		// TODO: check err return from updateCondition
		_ = r.updateCondition(
			builder.GetCtx(), builder.GetRequest(),
			ConditionAccessResourcesReady,
			metav1.ConditionFalse,
			"NotYetReady",
			fmt.Sprintf("%s", err))
		return err
	}

	return r.updateCondition(
		builder.GetCtx(), builder.GetRequest(),
		ConditionAccessResourcesReady,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		statusString)
}

// DeleteResource just deletes the resource immediately
//
// Returns:
//
//	error: Any error during the deletion
func (r *BaseRequestReconciler) DeleteResource(builder builders.IBuilder) error {
	return r.Delete(builder.GetCtx(), builder.GetRequest())
}
