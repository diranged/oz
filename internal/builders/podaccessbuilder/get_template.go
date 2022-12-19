package podaccessbuilder

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
)

// GetTemplate implements the IBuilder interface
func (b *PodAccessBuilder) GetTemplate(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
) (v1alpha1.ITemplateResource, error) {
	tmpl, err := req.GetTemplate(ctx, client)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, builders.ErrTemplateDoesNotExist
		}
		return nil, err
	}
	return tmpl, nil
}
