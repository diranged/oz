package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/diranged/oz/internal/metrics"
)

// recordAccessRequestCreated emits the create-time usage metrics for an access
// request. It is called from the validating webhook rather than the reconciler
// because the requesting user's identity is only available on the admission
// request.
//
// The namespace is taken from the admission request rather than the object,
// because the admission request is authoritative for namespaced resources even
// when the client omitted metadata.namespace.
//
// Dry-run admissions are skipped. The API server runs the full webhook chain
// for `kubectl apply --dry-run=server`, but nothing is persisted, so counting
// them would inflate usage with requests that were never actually granted.
func recordAccessRequestCreated(req admission.Request, obj IRequestResource) {
	if req.DryRun != nil && *req.DryRun {
		return
	}

	kind := metrics.Kind(obj)

	metrics.AccessRequestCreatedTotal.WithLabelValues(
		kind,
		req.Namespace,
		obj.GetTemplateName(),
		metrics.UserLabel(req.UserInfo.Username),
	).Inc()

	// Only observe a duration the user actually asked for. An unset (or
	// unparseable) spec.duration means the template's default applies, and we
	// have no client here with which to go resolve the template.
	if duration, err := obj.GetDuration(); err == nil && duration > 0 {
		metrics.AccessRequestRequestedDurationSeconds.WithLabelValues(
			kind,
			req.Namespace,
			obj.GetTemplateName(),
		).Observe(duration.Seconds())
	}
}
