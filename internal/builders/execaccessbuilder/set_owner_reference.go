package execaccessbuilder

import (
	"context"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetOwnerReference implements the IBuilder interface
func (b *ExecAccessBuilder) SetOwnerReference(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) error {
	return builders.SetOwnerReference(ctx, client, req, tmpl)
}
