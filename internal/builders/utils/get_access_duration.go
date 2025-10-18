package bldutil

import (
	"fmt"
	"time"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
)

// GetAccessDuration is a generic function for getting the proper Access
// Duration for a particular Access Request. This common logic can be shared
// across our IBuilders.
func GetAccessDuration(
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) (accessDuration time.Duration, decision string, err error) {
	// Step one - verify the inputs themselves. If the user supplied invalid inputs, or the template has any
	// invalid inputs, we bail out and update the conditions as such. This is to prevent escalated privilegess
	// from lasting indefinitely.
	var requestedDuration time.Duration
	if requestedDuration, err = req.GetDuration(); err != nil {
		return accessDuration, "", fmt.Errorf(
			"request error: %q: %w",
			builders.ErrRequestDurationInvalid,
			err,
		)
	}
	templateDefaultDuration, err := tmpl.GetAccessConfig().GetDefaultDuration()
	if err != nil {
		return accessDuration, "", fmt.Errorf(
			"template error: %q: %w",
			builders.ErrRequestDurationInvalid,
			err,
		)
	}
	templateMaxDuration, err := tmpl.GetAccessConfig().GetMaxDuration()
	if err != nil {
		return accessDuration, "", fmt.Errorf(
			"template error: %q: %w",
			builders.ErrRequestDurationInvalid,
			err,
		)
	}

	// Return the computed access duration
	accessDuration, decision = pickAccessDuration(
		requestedDuration,
		templateDefaultDuration,
		templateMaxDuration,
	)
	return accessDuration, decision, err
}

func pickAccessDuration(
	requestedDuration, defaultDuration, maxDuration time.Duration,
) (duration time.Duration, decision string) {
	if requestedDuration == 0 {
		// If no requested duration supplied, then default to the template's default duration
		duration = defaultDuration
		decision = fmt.Sprintf(
			"Access request duration defaulting to template duration time (%s)",
			defaultDuration.String(),
		)
	} else if requestedDuration <= maxDuration {
		// If the requested duration is too long, use the template max
		decision = fmt.Sprintf("Access requested custom duration (%s)", requestedDuration.String())
		duration = requestedDuration
	} else {
		// Finally, if it's valid, use the supplied duration
		decision = fmt.Sprintf(
			"Access requested duration (%s) larger than template maximum duration (%s)",
			requestedDuration.String(), maxDuration.String())
		duration = maxDuration
	}

	// Log out the decision, and update the condition
	return duration, decision
}
