package podwatcher

import (
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/diranged/oz/internal/metrics"
)

// recordPodConnect counts an exec or attach operation against a Pod.
//
// The Pod name is deliberately not recorded as a label - Oz-created Pods have
// generated names, so it would be unbounded cardinality. Use the Kubernetes
// Events that this webhook also emits when you need to trace a specific Pod.
func recordPodConnect(req admission.Request, interactive bool) {
	metrics.PodExecTotal.WithLabelValues(
		req.Namespace,
		req.SubResource,
		metrics.UserLabel(req.UserInfo.Username),
		strconv.FormatBool(interactive),
	).Inc()
}
