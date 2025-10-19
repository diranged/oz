package bldutil

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getDaemonSet returns a DaemonSet given the supplied generic client.Object resource
//
// Returns:
//
//	appsv1.DaemonSet: A populated deployment object
//	error: Any error that may have occurred
func getDaemonSet(
	ctx context.Context,
	client client.Client,
	obj client.Object,
) (*appsv1.DaemonSet, error) {
	found := &appsv1.DaemonSet{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)
	return found, err
}
