// Package metrics defines the custom Prometheus metrics that Oz exports
// alongside the standard controller-runtime metrics (reconcile counts, webhook
// latencies, workqueue depths, etc).
//
// The controller-runtime metrics tell you how hard the operator is working. The
// metrics in this package tell you how the operator is being *used*: who is
// requesting access, which templates they are requesting it through, how long
// those grants take to become usable, and how much exec/attach traffic the
// resulting Pods actually see.
//
// All metrics here are registered against the controller-runtime metrics
// registry, so they are served on the existing (authenticated) metrics endpoint
// without any extra wiring.
//
// # A note on the `user` label
//
// Attributing usage to a person is the whole point of several of these metrics,
// but usernames are unbounded-ish cardinality and, depending on your
// authenticator, are personally identifying (OIDC usernames are frequently
// email addresses). The `user` label is therefore opt-in: unless
// SetIncludeUserLabel(true) is called, every user label is reported as
// "redacted" so the series collapses to one per template. See the
// --metrics-include-user-label flag.
package metrics

import (
	"reflect"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// UserRedacted is reported in place of a real username when the `user`
	// label has not been explicitly enabled.
	UserRedacted = "redacted"

	// UserUnknown is reported when the Kubernetes API server did not supply
	// any user identity on the admission request. This should not happen in a
	// healthy cluster, and a non-trivial rate of it is worth alerting on.
	UserUnknown = "unknown"
)

// includeUserLabel gates whether real usernames are attached to metrics. It is
// read on every metric observation from webhook handlers, which run
// concurrently, hence the atomic.
var includeUserLabel atomic.Bool

// SetIncludeUserLabel enables or disables reporting of real usernames in the
// `user` label of the metrics in this package. It is expected to be called once
// during startup, before the manager begins serving webhooks.
func SetIncludeUserLabel(enabled bool) {
	includeUserLabel.Store(enabled)
}

// UserLabel maps a Kubernetes username onto a value safe to use as a metric
// label, honoring SetIncludeUserLabel.
//
// Note that a genuine user literally named "redacted" or "unknown" would be
// indistinguishable from the placeholders. This is not considered a problem in
// practice.
func UserLabel(username string) string {
	if !includeUserLabel.Load() {
		return UserRedacted
	}
	if username == "" {
		return UserUnknown
	}
	return username
}

// Kind returns the Kubernetes Kind for an Oz resource (eg
// "PodAccessRequest"), for use as a metric label.
//
// The concrete Go type name and the Kind are the same for every type in the
// v1alpha1 API, which lets callers label a metric without needing a runtime
// Scheme to do a GVK lookup.
func Kind(obj any) string {
	t := reflect.TypeOf(obj)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return "Unknown"
	}
	return t.Name()
}

// Reasons reported in the `reason` label of AccessRequestTerminatedTotal.
const (
	// ReasonExpired means the request outlived its access duration and was
	// deleted by the reconciler. This is the normal, happy-path ending for an
	// access request.
	ReasonExpired = "expired"

	// ReasonDurationInvalid means the request asked for a duration that could
	// not be parsed, and was abandoned without being retried.
	ReasonDurationInvalid = "duration_invalid"

	// ReasonDurationTooLong means the request asked for more time than its
	// template allows, and was abandoned without being retried.
	ReasonDurationTooLong = "duration_too_long"
)

var (
	// AccessRequestCreatedTotal counts access requests at the moment they are
	// admitted by the validating webhook. This is the headline "who is using
	// Oz, and through which template" metric.
	//
	// It counts admitted *attempts*: a request that is admitted here but then
	// rejected by a later admission plugin, or that fails to reconcile, is
	// still counted. Compare against AccessRequestReadySeconds's count to see
	// how many requests actually turned into working access.
	AccessRequestCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oz_access_request_created_total",
			Help: "Total number of Oz access requests admitted by the validating webhook, by kind, namespace, template and requesting user.",
		},
		[]string{"kind", "namespace", "template", "user"},
	)

	// AccessRequestRequestedDurationSeconds observes the duration a user
	// explicitly asked for on creation.
	//
	// Requests that omit spec.duration (and therefore inherit the template's
	// default) are NOT observed here, because the webhook has no client with
	// which to resolve the template. A large gap between this metric's count
	// and AccessRequestCreatedTotal means most users are taking the default.
	AccessRequestRequestedDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "oz_access_request_requested_duration_seconds",
			Help: "Access duration explicitly requested in spec.duration, in seconds. Requests that omit spec.duration and inherit the template default are not observed.",
			Buckets: []float64{
				900,    // 15m
				1800,   // 30m
				3600,   // 1h
				7200,   // 2h
				14400,  // 4h
				28800,  // 8h
				43200,  // 12h
				86400,  // 24h
				172800, // 48h
			},
		},
		[]string{"kind", "namespace", "template"},
	)

	// AccessRequestReadySeconds observes how long it took a request to go from
	// creation to Status.Ready, ie how long the user waited before they could
	// actually use their access. For PodAccessRequests this includes cloning
	// the target workload's PodSpec and waiting for the new Pod to start, so
	// it is the metric to watch if users complain that Oz is slow.
	//
	// Observed once per request, on the transition to ready.
	AccessRequestReadySeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "oz_access_request_ready_seconds",
			Help:    "Time from access request creation to Status.Ready becoming true, in seconds.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10),
		},
		[]string{"kind", "namespace", "template"},
	)

	// AccessRequestTerminatedTotal counts requests that reached a terminal
	// state, labelled by why. See the Reason* constants; `expired` is the
	// happy path, everything else is a user or configuration error.
	AccessRequestTerminatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oz_access_request_terminated_total",
			Help: "Total number of Oz access requests that reached a terminal state, by reason (expired, duration_invalid, duration_too_long).",
		},
		[]string{"kind", "namespace", "template", "reason"},
	)

	// AccessRequestConditionErrorsTotal counts reconcile attempts that failed
	// a verification step, labelled with the Status.Condition that was set to
	// False.
	//
	// This is per *attempt*, not per request: a request wedged on a missing
	// template increments this on every reconciliation interval. That is
	// deliberate - it turns an otherwise silent requeue loop into a rate you
	// can alert on.
	AccessRequestConditionErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oz_access_request_condition_errors_total",
			Help: "Total number of access request reconcile attempts that failed a verification step, by the Status.Condition set to False.",
		},
		[]string{"kind", "namespace", "template", "condition"},
	)

	// PodExecTotal counts exec and attach operations seen by the Oz pod
	// watcher webhook.
	//
	// This is the "did the access actually get used" counter. Note that the
	// webhook is registered for pods/exec and pods/attach cluster-wide, not
	// just for Oz-managed Pods, so this metric covers all exec traffic in the
	// cluster. The Pod name is deliberately not a label - it is unbounded.
	PodExecTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oz_pod_exec_total",
			Help: "Total number of Pod exec/attach operations observed by the Oz pod watcher, by namespace, subresource, user and whether a TTY was requested.",
		},
		[]string{"namespace", "subresource", "user", "interactive"},
	)
)

func init() {
	metrics.Registry.MustRegister(
		AccessRequestCreatedTotal,
		AccessRequestRequestedDurationSeconds,
		AccessRequestReadySeconds,
		AccessRequestTerminatedTotal,
		AccessRequestConditionErrorsTotal,
		PodExecTotal,
	)
}
