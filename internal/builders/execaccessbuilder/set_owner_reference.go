package execaccessbuilder

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// SetOwnerReference implements the IBuilder interface
func (b *ExecAccessBuilder) SetOwnerReference(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
) error {
	// Get the template object... this should be fast, it's likely cached by now.
	tmpl, err := req.GetTemplate(ctx, client)
	if err != nil {
		return err
	}
	// Set the controller owner reference
	if err := ctrl.SetControllerReference(tmpl, req, client.Scheme()); err != nil {
		return err
	}
	// Push the update back to K8S
	return client.Update(ctx, req)
}
