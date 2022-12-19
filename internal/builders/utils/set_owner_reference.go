package utils

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetOwnerReference provides a generic wrapper for setting the OwnerReference
// on a resource and updating the pointer to that resource. This function is
// used by the individual builders to implement the IBuilder interface.
func SetOwnerReference(
	ctx context.Context,
	client client.Client,
	owner client.Object,
	controlled client.Object,
) error {
	// Set the controller owner reference
	if err := ctrl.SetControllerReference(owner, controlled, client.Scheme()); err != nil {
		return err
	}
	// Push the update back to K8S
	return client.Update(ctx, controlled)
}
