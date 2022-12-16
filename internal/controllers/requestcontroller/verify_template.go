package requestcontroller

import (
	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/controllers/internal/status"
)

// verifyTemplate asks the IBuilder object to verify that the target template
// exists. We update the request condition accordingly and return. Any error
// returned should trigger the end of the reconciliation loop and for it to
// requeue. The template object itself is returned and used for the next stages
// in the reconciliation.
func (r *RequestReconciler) verifyTemplate(
	rctx *RequestContext,
) (v1alpha1.ITemplateResource, error) {
	tmpl, err := r.Builder.GetTemplate(rctx.Context, r.Client, rctx.obj)
	if err != nil {
		rctx.log.Error(err, "Unable to verify template")

		// Update the condition. If that fails, return the error, otherwise
		// return nil which continues reconciliation.
		if err := status.SetTargetTemplateNotExists(rctx.Context, r, rctx.obj, err); err != nil {
			return nil, err
		}

		// Return the original error now to fail reconciliation.
		return nil, err
	}

	// Update the condition and return. Any failure on updating this condition
	// will fail reconciliation.
	if err := status.SetTargetTemplateExists(rctx.Context, r, rctx.obj); err != nil {
		return nil, err
	}

	// UPDATE: Set the OwnerReference for the request - so if the template is
	// deleted, all requests are deleted.
	if err := r.Builder.SetOwnerReference(rctx.Context, r.Client, rctx.obj, tmpl); err != nil {
		rctx.log.Error(err, "Error setting owner reference")
		return nil, err
	}

	return tmpl, nil
}
