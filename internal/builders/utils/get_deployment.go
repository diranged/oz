package bldutil

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getDeployment returns a Deployment given the supplied generic client.Object resource
//
// Returns:
//
//	appsv1.Deployment: A populated deployment object
//	error: Any error that may have occurred
func getDeployment(
	ctx context.Context,
	client client.Client,
	obj client.Object,
) (*appsv1.Deployment, error) {
	found := &appsv1.Deployment{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)
	return found, err
}
