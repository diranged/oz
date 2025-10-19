package podaccessbuilder

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
	bldutil "github.com/diranged/oz/internal/builders/utils"
)

// SetRequestOwnerReference implements the IBuilder interface
func (b *PodAccessBuilder) SetRequestOwnerReference(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) error {
	return bldutil.SetOwnerReference(ctx, client, tmpl, req)
}
