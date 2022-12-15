package request_controller

import (
	"context"

	"github.com/diranged/oz/internal/controllers"
	"github.com/diranged/oz/internal/controllers/internal/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetupWithManager sets up the controller with the Manager.
func (r *RequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Provide a searchable index in the cached kubernetes client for "metadata.name" - the pod name.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, controllers.FieldSelectorMetadataName, func(rawObj client.Object) []string {
		// grab the job object, extract the name...
		pod := rawObj.(*v1.Pod)
		name := pod.GetName()
		return []string{name}
	}); err != nil {
		return err
	}

	// Provide a searchable index in the cached kubernetes client for "status.phase", allowing us to
	// search for Running Pods.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, controllers.FieldSelectorStatusPhase, func(rawObj client.Object) []string {
		// grab the job object, extract the phase...
		pod := rawObj.(*v1.Pod)
		phase := string(pod.Status.Phase)
		return []string{phase}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(r.RequestType).
		WithEventFilter(utils.IgnoreStatusUpdatesAndDeletion()).
		Complete(r)
}
