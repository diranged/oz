package v1alpha1

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/diranged/oz/internal/metrics"
)

// histogramSumFor returns the observed sum for one label set of the requested
// duration histogram. testutil.ToFloat64 cannot be used here - it only handles
// single-value metrics, not histograms - so gather the vector into a scratch
// registry and pick out the matching series.
func histogramSumFor(labelValues ...string) float64 {
	registry := prometheus.NewPedanticRegistry()
	ExpectWithOffset(1, registry.Register(metrics.AccessRequestRequestedDurationSeconds)).
		To(Succeed())

	families, err := registry.Gather()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	want := strings.Join(labelValues, "\x00")
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			got := []string{}
			for _, label := range metric.GetLabel() {
				got = append(got, label.GetValue())
			}
			if strings.Join(got, "\x00") == want {
				return metric.GetHistogram().GetSampleSum()
			}
		}
	}
	return -1
}

var _ = Describe("Webhook metrics", func() {
	const (
		namespace = "metrics-test"
		template  = "webhook-metrics-template"
	)

	// createRequestFor builds the admission request the API server would send
	// for a create, optionally flagged as a dry run.
	createRequestFor := func(dryRun bool) admission.Request {
		return admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Namespace: namespace,
				DryRun:    &dryRun,
				UserInfo:  authenticationv1.UserInfo{Username: "dennis"},
			},
		}
	}

	created := func() float64 {
		return testutil.ToFloat64(metrics.AccessRequestCreatedTotal.WithLabelValues(
			"PodAccessRequest", namespace, template, metrics.UserRedacted,
		))
	}

	It("counts a real create", func() {
		before := created()
		recordAccessRequestCreated(
			createRequestFor(false),
			&PodAccessRequest{Spec: PodAccessRequestSpec{TemplateName: template}},
		)
		Expect(created()).To(Equal(before + 1))
	})

	// The API server runs the whole webhook chain for `kubectl apply
	// --dry-run=server`, but persists nothing. Counting those would report
	// access that was never granted.
	It("does not count a dry-run create", func() {
		before := created()
		recordAccessRequestCreated(
			createRequestFor(true),
			&PodAccessRequest{Spec: PodAccessRequestSpec{TemplateName: template}},
		)
		Expect(created()).To(Equal(before))
	})

	It("does not count a dry-run create for ExecAccessRequests either", func() {
		labels := []string{"ExecAccessRequest", namespace, template, metrics.UserRedacted}
		before := testutil.ToFloat64(
			metrics.AccessRequestCreatedTotal.WithLabelValues(labels...),
		)
		recordAccessRequestCreated(
			createRequestFor(true),
			&ExecAccessRequest{Spec: ExecAccessRequestSpec{TemplateName: template}},
		)
		Expect(testutil.ToFloat64(
			metrics.AccessRequestCreatedTotal.WithLabelValues(labels...),
		)).To(Equal(before))
	})

	It("observes an explicitly requested duration", func() {
		before := testutil.CollectAndCount(metrics.AccessRequestRequestedDurationSeconds)

		recordAccessRequestCreated(
			createRequestFor(false),
			&PodAccessRequest{Spec: PodAccessRequestSpec{
				TemplateName: "duration-template",
				Duration:     "4h",
			}},
		)

		Expect(testutil.CollectAndCount(metrics.AccessRequestRequestedDurationSeconds)).
			To(BeNumerically(">", before))
		// Labels are gathered in alphabetical order: kind, namespace, template.
		Expect(histogramSumFor("PodAccessRequest", namespace, "duration-template")).
			To(Equal(float64(4 * 60 * 60)))
	})

	// An omitted spec.duration means "use the template default", which the
	// webhook cannot resolve, so there is nothing meaningful to observe.
	It("does not observe a request that inherits the template default", func() {
		before := testutil.CollectAndCount(metrics.AccessRequestRequestedDurationSeconds)
		recordAccessRequestCreated(
			createRequestFor(false),
			&PodAccessRequest{Spec: PodAccessRequestSpec{TemplateName: "no-duration-template"}},
		)
		Expect(testutil.CollectAndCount(metrics.AccessRequestRequestedDurationSeconds)).
			To(Equal(before))
	})
})
