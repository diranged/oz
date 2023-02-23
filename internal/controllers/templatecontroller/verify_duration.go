package templatecontroller

import (
	"fmt"

	"github.com/diranged/oz/internal/controllers/internal/status"
)

// verifyDuration walks through the AccessConfig settings for an
// ITemplateResource and verifies that the inputs are sane. Conditions are
// updated if they are not, but errors are only returned if the condition
// update process fails.
func (r *TemplateReconciler) verifyDuration(rctx *RequestContext) error {
	eventStr := "DurationVerified"

	// Verify that MaxDuration is greater than DesiredDuration.
	defaultDuration, err := rctx.obj.GetAccessConfig().GetDefaultDuration()
	if err != nil {
		return status.SetTemplateDurationsNotValid(rctx.Context, r, rctx.obj,
			fmt.Sprintf("Error on spec.defaultDuration: %s", err),
		)
	}
	maxDuration, err := rctx.obj.GetAccessConfig().GetMaxDuration()
	if err != nil {
		errStr := fmt.Sprintf("Error on spec.maxDuration: %s", err)
		r.recorder.Event(rctx.obj, "Warning", eventStr, errStr)
		return status.SetTemplateDurationsNotValid(rctx.Context, r, rctx.obj, errStr)
	}
	if defaultDuration > maxDuration {
		errStr := "Error: spec.defaultDuration can not be greater than spec.maxDuration"
		r.recorder.Event(rctx.obj, "Warning", eventStr, errStr)
		return status.SetTemplateDurationsNotValid(rctx.Context, r, rctx.obj, errStr)
	}

	successStr := "spec.defaultDuration and spec.maxDuration valid"
	r.recorder.Event(rctx.obj, "Normal", eventStr, successStr)
	return status.SetTemplateDurationsValid(rctx.Context, r, rctx.obj, successStr)
}
