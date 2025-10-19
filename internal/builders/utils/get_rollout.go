package bldutil

import (
	"context"

	rolloutsv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getRollout returns a Rollout given the supplied generic client.Object resource
//
// Returns:
//
//	v1alpha1.Rollout: A populated Rollout object
//	error: Any error that may have occurred
func getRollout(
	ctx context.Context,
	client client.Client,
	obj client.Object,
) (*rolloutsv1alpha1.Rollout, error) {
	found := &rolloutsv1alpha1.Rollout{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)

	return found, err
}
