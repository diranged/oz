package controllers

import (
	"fmt"
	"time"

	"github.com/diranged/oz/controllers/builders"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OzRequestReconciler provides a base reconciler with common functions for handling our Template CRDs
// (ExecAccessTemplate, AccessTemplate, etc)
type OzRequestReconciler struct {
	*OzReconciler
}

// VerifyDuration checks a few components of whether or not the AccessRequest is still valid:
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
func (r *OzRequestReconciler) VerifyDuration(builder builders.Builder) error {
	var err error
	logger := r.GetLogger(builder.GetCtx())

	logger.Info("Beginning access request duration verification")

	// Step one - verify the inputs themselves. If the user supplied invalid inputs, or the template has any
	// invalid inputs, we bail out and update the conditions as such. This is to prevent escalated privilegess
	// from lasting indefinitely.
	var requestedDuration time.Duration
	if requestedDuration, err = builder.GetRequest().GetDuration(); err != nil {
		r.UpdateCondition(builder.GetCtx(), builder.GetRequest(), ConditionDurationsValid,
			metav1.ConditionFalse, string(metav1.StatusReasonBadRequest), fmt.Sprintf("spec.duration error: %s", err))
		return err
	}
	templateDefaultDuration, err := builder.GetTemplate().GetDefaultDuration()
	if err != nil {
		r.UpdateCondition(builder.GetCtx(), builder.GetRequest(), ConditionDurationsValid,
			metav1.ConditionFalse, string(metav1.StatusReasonBadRequest), fmt.Sprintf("Template Error, spec.defaultDuration error: %s", err))
		return err
	}

	templateMaxDuration, err := builder.GetTemplate().GetMaxDuration()
	if err != nil {
		r.UpdateCondition(builder.GetCtx(), builder.GetRequest(), ConditionDurationsValid,
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

	err = r.UpdateCondition(builder.GetCtx(), builder.GetRequest(), ConditionDurationsValid,
		metav1.ConditionTrue, string(metav1.StatusSuccess), reasonStr)
	if err != nil {
		return err
	}

	// If the accessUptime is greater than the accessDuration, kill it.
	if builder.GetRequest().GetUptime() > accessDuration {
		return r.UpdateCondition(builder.GetCtx(), builder.GetRequest(), ConditionAccessStillValid,
			metav1.ConditionFalse, string(metav1.StatusReasonTimeout), "Access expired")
	}

	// Update the resource, and let the user know how much time is remaining
	return r.UpdateCondition(builder.GetCtx(), builder.GetRequest(), ConditionAccessStillValid,
		metav1.ConditionTrue, string(metav1.StatusReasonTimeout),
		"Access still valid")
}

// IsAccessExpired checks the AccessRequest status for the ConditionAccessStillValid condition. If it is no longer
// a valid request, then the resource is immediately deleted.
//
// Returns:
//
//	true: if the resource is expired, AND has now been deleted
//	false: if the resource is still valid
//	error: any error during the checks
func (r *OzRequestReconciler) IsAccessExpired(builder builders.Builder) (bool, error) {
	logger := r.GetLogger(builder.GetCtx())
	logger.Info("Checking if access has expired or not...")
	cond := meta.FindStatusCondition(*builder.GetRequest().GetConditions(), string(ConditionAccessStillValid))
	if cond == nil {
		logger.Info(fmt.Sprintf("Missing Condition %s, skipping deletion", ConditionAccessStillValid))
		return false, nil
	}

	if cond.Status == metav1.ConditionFalse {
		logger.Info(fmt.Sprintf("Found Condition %s in state %s, terminating rqeuest", ConditionAccessStillValid, cond.Status))
		return true, r.DeleteResource(builder)
	}

	logger.Info(fmt.Sprintf("Found Condition %s in state %s, leaving alone", ConditionAccessStillValid, cond.Status))
	return false, nil
}

// GetPodName returns back the name of a the pod that this Request will grant access to. If no
// existing Podname exists on the RequestAccess resource, then the GeneratePodname() function is
// called. The GeneratePodname() function needs to be implemented uniquely in each Builder interface.
//
// Returns:
//
//	podname: A string reference to the individual Pod name
//	error: Any error either getting the podName or updating the condition of the access request.
func (r *OzRequestReconciler) GetPodName(builder builders.Builder) (string, error) {
	ctx := builder.GetCtx()
	request := builder.GetRequest()
	var podName string
	var err error

	logger := log.FromContext(builder.GetCtx())

	// If this resource already has a status.podName field set, then we respect that no matter what.
	// We never mutate the pod that this access request was originally created for. Otherwise, pick
	// a Pod and populate that status field.
	if builder.GetRequest().GetPodName() != "" {
		logger.Info(fmt.Sprintf("Pod already assigned - %s", builder.GetRequest().GetPodName()))
		return builder.GetRequest().GetPodName(), nil
	}

	// If the GeneratePodName() function fails for any reason, we bail out AND wipe out the condition
	// in case it had previously been set.
	if podName, err = builder.GeneratePodName(); err != nil {
		r.UpdateCondition(
			ctx, request,
			ConditionTargetPodSelected,
			metav1.ConditionFalse,
			string(metav1.StatusFailure),
			fmt.Sprintf("ERROR: %s", err))
		return "", err
	}

	// Set the podName (note, just in the local object). If this fails (for example, its already set
	// on the object), then we also bail out.
	if err := request.SetPodName(podName); err != nil {
		return "", err
	}

	// Finally, update the condition with the podname - this has the side effect of pushing the
	// value of Status.PodName into Kubernetes.
	if err := r.UpdateCondition(
		ctx, request,
		ConditionTargetPodSelected,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		fmt.Sprintf("Pod %s selected", request.GetPodName())); err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("Target Pod Name %s", builder.GetRequest().GetPodName()))
	return builder.GetRequest().GetPodName(), nil
}

// DeleteResource just deletes the resource immediately
//
// Returns:
//
//	error: Any error during the deletion
func (r *OzRequestReconciler) DeleteResource(builder builders.Builder) error {
	return r.Delete(builder.GetCtx(), builder.GetRequest())
}
