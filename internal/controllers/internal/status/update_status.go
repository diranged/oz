package status

import (
	"context"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/internal/api/v1alpha1"
	ctrlutil "github.com/diranged/oz/internal/controllers/internal/utils"
)

// UpdateStatus pushes the client.Object.Status field into Kubernetes if it has been updated, and
// then takes care of calling Refetch() to re-populate the object pointer with the updated object
// revision from Kubernetes.
//
// This wrapper makes it much easier to update the Status field of an object iteratively throughout
// a reconciliation loop.
func UpdateStatus(ctx context.Context, rec hasStatusReconciler, res api.ICoreResource) error {
	log := logf.FromContext(ctx)

	// Update the status, handle failure.
	if err := rec.Status().Update(ctx, res); err != nil {
		log.Error(err, "Failed to update status")
		return err
	}

	// Re-fetch the object when we're done to make sure we are working with the latest version
	if _, err := ctrlutil.Refetch(ctx, rec.GetAPIReader(), res); err != nil {
		log.Error(err, "Failed to refetch object")
		return err
	}

	return nil
}
