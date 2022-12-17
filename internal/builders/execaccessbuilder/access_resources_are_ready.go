package execaccessbuilder

import (
	"context"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AccessResourcesAreReady implements the IBuilder interface
func (b *ExecAccessBuilder) AccessResourcesAreReady(
	_ context.Context,
	_ client.Client,
	_ v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) (bool, error) {
	// There is no waiting for resources to come up here. Everything we create
	// is automatically available.
	return true, nil
}
