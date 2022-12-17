package requestcontroller

import (
	"errors"
	"time"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"github.com/diranged/oz/internal/controllers/internal/ctrlrequeue"
	"github.com/diranged/oz/internal/controllers/internal/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *RequestReconciler) verifyDuration(
	rctx *RequestContext,
	tmpl v1alpha1.ITemplateResource,
) (shouldEndReconcile bool, result ctrl.Result, resultErr error) {
	var accessDuration time.Duration
	var decision string

	rctx.log.V(1).Info("Computing Access Request duration...")

	// Get the accessDuration and decision from the builder
	accessDuration, decision, err := r.Builder.GetAccessDuration(rctx.obj, tmpl)
	// If an error is returned, determine whether its something wrong with the
	// user-supplied inputs, or whether it was transient.
	if err != nil {
		switch errors.Unwrap(err) {
		case builders.ErrRequestDurationInvalid:
			rctx.log.Error(err, "RequestDurationInvalid, will not requeue.")
			shouldEndReconcile = true
			result, resultErr = ctrlrequeue.NoRequeue()
		case builders.ErrRequestDurationTooLong:
			rctx.log.Error(err, "RequestDurationTooLong, will not requeue.")
			shouldEndReconcile = true
			result, resultErr = ctrlrequeue.NoRequeue()
		default:
			rctx.log.Error(err, "Unexpected error, will requeue")
			shouldEndReconcile = true
			result, resultErr = ctrlrequeue.RequeueError(err)
		}

		// Update the status, and return the results
		_ = status.SetRequestDurationsNotValid(rctx.Context, r, rctx.obj, err.Error())
		return shouldEndReconcile, result, resultErr
	}

	// Success, update the resource
	if err := status.SetRequestDurationsValid(rctx.Context, r, rctx.obj, decision); err != nil {
		return true, ctrl.Result{}, err
	}

	// If the access is expired at this point, update that condition too.
	if rctx.obj.GetUptime() > accessDuration {
		// No we should not end the reconcile - the access is invalid ... but
		// that means we need to finish the reconcile to trigger the deletion
		// phase. Only requeue if the SetAccessNotValid() step fails.
		return false, result, status.SetAccessNotValid(rctx.Context, r, rctx.obj)
	}

	// End by setting the access to still-valid
	return false, result, status.SetAccessStillValid(rctx.Context, r, rctx.obj)
}
