package requestcontroller

import (
	ctrlutil "github.com/diranged/oz/internal/controllers/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupWithManager sets up the controller with the Manager.
func (r *RequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.RequestType).
		WithEventFilter(ctrlutil.IgnoreStatusUpdatesAndDeletion()).
		Complete(r)
}
