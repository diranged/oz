package request_controller

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"github.com/diranged/oz/internal/controllers/internal/status"
)

// verifyTemplate asks the IBuilder object to verify that the target template
// exists. We update the request condition accordingly and return. Any error
// returned should trigger the end of the reconciliation loop and for it to
// requeue.
func (r *RequestReconciler) verifyTemplate(
	rctx *RequestContext,
) (v1alpha1.ITemplateResource, error) {
	tmpl, err := r.Builder.GetTemplate(rctx.Context, r.Client, rctx.obj)
	if err != nil {
		rctx.log.Error(err, "Unable to verify template")

		// Attempt to update the condition with a friendly-enough error message
		// about the template being incorrect.
		switch {
		case errors.Is(err, builders.ErrTemplateDoesNotExist):
			// Update the condition. If that fails, return the error, otherwise
			// return nil which continues reconciliation.
			if err := status.UpdateCondition(
				rctx.Context, r, rctx.obj,
				v1alpha1.ConditionTargetTemplateExists,
				metav1.ConditionFalse,
				"TemplateNotFound",
				fmt.Sprintf("Error: %s", err)); err != nil {
				return nil, err
			}
		default:
			// Update the condition. If this fails, return that error which will
			// fail reconciliation.
			if err := status.UpdateCondition(
				rctx.Context, r, rctx.obj,
				v1alpha1.ConditionTargetTemplateExists,
				metav1.ConditionFalse,
				"Unknown",
				fmt.Sprintf("Error: %s", err)); err != nil {
				return nil, err
			}
		}

		// Return the original error now to fail reconciliation.
		return nil, err
	}

	// Update the condition and return. Any failure on updating this condition
	// will fail reconciliation.
	if err := status.UpdateCondition(
		rctx.Context, r, rctx.obj,
		v1alpha1.ConditionTargetTemplateExists,
		metav1.ConditionTrue,
		string(metav1.StatusSuccess),
		"Found Target Template",
	); err != nil {
		return nil, err
	}

	return tmpl, nil
}
