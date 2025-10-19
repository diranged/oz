package bldutil

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getStatefulSet returns a StatefulSet given the supplied generic client.Object resource
//
// Returns:
//
//	appsv1.StatefulSet: A populated deployment object
//	error: Any error that may have occurred
func getStatefulSet(
	ctx context.Context,
	client client.Client,
	obj client.Object,
) (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)
	return found, err
}
