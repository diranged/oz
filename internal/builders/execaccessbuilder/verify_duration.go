package execaccessbuilder

import (
	"fmt"
	"time"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
)

// VerifyDuration implements the IBuilder interface
func (b *ExecAccessBuilder) VerifyDuration(
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) (expiresAt time.Time, err error) {
	// Step one - if the request already has an expiration time, we honor that and move on.
	//	if req.GetStatus()

	// Step one - verify the inputs themselves. If the user supplied invalid inputs, or the template has any
	// invalid inputs, we bail out and update the conditions as such. This is to prevent escalated privilegess
	// from lasting indefinitely.
	var requestedDuration time.Duration
	if requestedDuration, err = req.GetDuration(); err != nil {
		return "", fmt.Errorf("request error: %w: %w", builders.ErrRequestDurationInvalid, err)
	}
	templateDefaultDuration, err := tmpl.GetAccessConfig().GetDefaultDuration()
	if err != nil {
		return "", fmt.Errorf("template error: %w: %w", builders.ErrRequestDurationInvalid, err)
	}
	templateMaxDuration, err := tmpl.GetAccessConfig().GetMaxDuration()
	if err != nil {
		return "", fmt.Errorf("template error: %w: %w", builders.ErrRequestDurationInvalid, err)
	}

	// Now determine which duration is the one we'll use
	accessDuration := b.getAccessDuration(
		requestedDuration,
		templateDefaultDuration,
		templateMaxDuration,
	)

	if req.GetUptime() > accessDuration {
		return "", builders.ErrRequestExpired
	}
}

func (b *ExecAccessBuilder) getAccessDuration(
	requestedDuration, defaultDuration, maxDuration time.Duration,
) (accessDuration time.Duration) {
	if requestedDuration == 0 {
		// If no requested duration supplied, then default to the template's default duration
		//reason = fmt.Sprintf(
		//	"Access request duration defaulting to template duration time (%s)",
		//	defaultDuration.String(),
		//)
		accessDuration = defaultDuration
	} else if requestedDuration <= maxDuration {
		// If the requested duration is too long, use the template max
		// reason = fmt.Sprintf("Access requested custom duration (%s)", requestedDuration.String())
		accessDuration = requestedDuration
	} else {
		// Finally, if it's valid, use the supplied duration
		// reason = fmt.Sprintf(
		//	"Access requested duration (%s) larger than template maximum duration (%s)",
		//	requestedDuration.String(), maxDuration.String())
		accessDuration = maxDuration
	}

	// Log out the decision, and update the condition
	return accessDuration, reason
}
