package podaccessbuilder

import (
	"time"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders/utils"
)

// GetAccessDuration implements the IBuilder interface
func (b *PodAccessBuilder) GetAccessDuration(
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) (time.Duration, string, error) {
	return utils.GetAccessDuration(req, tmpl)
}
