package requestcontroller

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// fetchRequestObject fetches the Kubernetes API object for the Component that
// this reconcile is running for.
func (r *RequestReconciler) fetchRequestObject(rctx *RequestContext) error {
	log := log.FromContext(rctx.Context)
	err := r.Get(rctx.Context, rctx.req.NamespacedName, rctx.obj)
	if err != nil {
		log.V(3).Info(fmt.Sprintf("%s not found: %s", rctx.obj.GetObjectKind(), err.Error()))
	}
	return err
}
