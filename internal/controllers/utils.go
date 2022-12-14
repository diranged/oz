package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// refetch uses the "consistent client" (non-caching) to retreive the latest state of the object into the
// supplied object reference. This is critical to avoid "the object has been modified; please apply
// your changes to the latest version and try again" errors when updating object status fields.
func refetch(ctx context.Context, reader client.Reader, obj client.Object) (*client.Object, error) {
	if err := reader.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// ignoreStatusUpdatesAndDeletion filters out reconcile requests where only the
// Status was updated, or on Deletes.
//
// **Deletes**
// On Deletes, we don't need to do any cleanup because we make sure to use
// OwnerReferences that force Kubernetes to handle the cleanup for us.
//
// **Status Updates**
// Our Reconcile() loops make many updates mid-reconcile to the status fields
// of the objects. Doing this can cause all kinds of re-runs of the reconciler
// at a high rate - mostly when they are not desired.
//
// Using this predicate filter means that the Reconcile() loops must be well
// tested and include their own automatic requeue-after settings.
//
// https://sdk.operatorframework.io/docs/building-operators/golang/references/event-filtering/
func ignoreStatusUpdatesAndDeletion() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}
