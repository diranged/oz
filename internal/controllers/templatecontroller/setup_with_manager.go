package templatecontroller

import (
	ctrlutil "github.com/diranged/oz/internal/controllers/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupWithManager sets up the controller with the Manager.
func (r *TemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.TemplateType).
		WithEventFilter(ctrlutil.IgnoreStatusUpdatesAndDeletion()).
		Complete(r)
}
