package builders

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// IBuilder defines an interface that our RequestController can use to manage Access Request resources
type IBuilder interface {
	// VerifyTemplate checks whether or not the TargetTemplate actually exists
	VerifyTemplate(ctx context.Context, client client.Client, req v1alpha1.IRequestResource) error

	// SetOwnerReference ensures that if the TargetTemplate is ever deleted,
	// that all of the Access Requests pointing to it are also automatically
	// deleted, which automatically cascades down to delete all of the access
	// resources.
	//
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	SetOwnerReference(
		ctx context.Context,
		client client.Client,
		req v1alpha1.IRequestResource,
	) error
}
