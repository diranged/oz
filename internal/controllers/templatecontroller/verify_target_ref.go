package templatecontroller

import (
	"github.com/diranged/oz/internal/controllers/internal/status"
	"k8s.io/apimachinery/pkg/types"
)

// verifyTargetRef ensures that the Spec.targetRef points to a valid and
// understood controller that we can build our templates off of. Any failure
// results in the resource ConditionTargetRefExists condition being set to
// False.
//
// Returns:
//   - An "error" only if the UpdateCondition function fails
func (r *TemplateReconciler) verifyTargetRef(rctx *RequestContext) error {
	rctx.log.Info("Beginning TargetRef Verification")

	// https://blog.gripdev.xyz/2020/07/20/k8s-operator-with-dynamic-crds-using-controller-runtime-no-structs/
	targetRef := rctx.obj.GetTargetRef().GetObject()
	err := r.Client.Get(rctx.Context, types.NamespacedName{
		Name:      rctx.obj.GetTargetRef().GetName(),
		Namespace: rctx.obj.GetNamespace(),
	}, targetRef)
	if err != nil {
		return status.SetTargetRefNotExists(rctx.Context, r, rctx.obj, err)
	}
	return status.SetTargetRefExists(rctx.Context, r, rctx.obj, "Success")
}
