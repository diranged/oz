package execaccessbuilder

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
)

// VerifyTemplate implements the IBuilder interface
func (b *ExecAccessBuilder) VerifyTemplate(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
) error {
	_, err := req.GetTemplate(ctx, client)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return builders.ErrTemplateDoesNotExist
		}
		return err
	}
	return nil
}
