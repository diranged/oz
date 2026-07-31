package requestcontroller

import (
	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/metrics"
)

// metricLabels returns the (kind, namespace, template) label triple shared by
// the access request metrics. Safe to call any time after fetchRequestObject
// has populated rctx.obj.
func (rctx *RequestContext) metricLabels() (kind, namespace, template string) {
	return metrics.Kind(rctx.obj), rctx.obj.GetNamespace(), rctx.obj.GetTemplateName()
}

// recordConditionError counts a reconcile attempt that failed a verification
// step. Called once per failing attempt, so a permanently wedged request keeps
// incrementing this on every reconciliation interval.
func recordConditionError(rctx *RequestContext, condition v1alpha1.IConditionType) {
	kind, namespace, template := rctx.metricLabels()
	metrics.AccessRequestConditionErrorsTotal.
		WithLabelValues(kind, namespace, template, condition.String()).
		Inc()
}

// recordTerminated counts a request reaching a terminal state. See the
// metrics.Reason* constants.
func recordTerminated(rctx *RequestContext, reason string) {
	kind, namespace, template := rctx.metricLabels()
	metrics.AccessRequestTerminatedTotal.
		WithLabelValues(kind, namespace, template, reason).
		Inc()
}

// recordReady observes how long the requesting user waited between creating
// their request and it becoming usable. Callers must only invoke this on the
// transition into the ready state, so that each request is observed once.
func recordReady(rctx *RequestContext) {
	kind, namespace, template := rctx.metricLabels()
	metrics.AccessRequestReadySeconds.
		WithLabelValues(kind, namespace, template).
		Observe(rctx.obj.GetUptime().Seconds())
}
