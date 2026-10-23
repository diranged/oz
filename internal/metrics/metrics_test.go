package metrics

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

var _ = Describe("Metrics", func() {
	// The user label gate is global state, so every test that flips it must put
	// it back afterwards.
	AfterEach(func() {
		SetIncludeUserLabel(false)
	})

	Context("UserLabel", func() {
		It("redacts usernames by default", func() {
			Expect(UserLabel("dennis@example.com")).To(Equal(UserRedacted))
		})

		It("redacts an empty username too, rather than leaking the distinction", func() {
			Expect(UserLabel("")).To(Equal(UserRedacted))
		})

		It("returns the real username once enabled", func() {
			SetIncludeUserLabel(true)
			Expect(UserLabel("dennis@example.com")).To(Equal("dennis@example.com"))
		})

		It("reports a missing identity as unknown once enabled", func() {
			SetIncludeUserLabel(true)
			Expect(UserLabel("")).To(Equal(UserUnknown))
		})
	})

	Context("Kind", func() {
		type somethingElse struct{}

		It("returns the type name for a pointer", func() {
			Expect(Kind(&somethingElse{})).To(Equal("somethingElse"))
		})

		It("returns the type name for a value", func() {
			Expect(Kind(somethingElse{})).To(Equal("somethingElse"))
		})

		It("handles a nil interface without panicking", func() {
			Expect(Kind(nil)).To(Equal("Unknown"))
		})
	})

	Context("Registration", func() {
		// A duplicate metric name, or a Help string that disagrees between two
		// registrations, would panic at init() time. Confirm the metrics are
		// actually usable and land where we expect.
		It("registers the metrics against the controller-runtime registry", func() {
			AccessRequestCreatedTotal.
				WithLabelValues("PodAccessRequest", "ns", "tmpl", UserRedacted).
				Inc()

			Expect(testutil.ToFloat64(
				AccessRequestCreatedTotal.WithLabelValues(
					"PodAccessRequest", "ns", "tmpl", UserRedacted,
				),
			)).To(BeNumerically(">=", 1))
		})

		It("exposes a histogram per template for grant latency", func() {
			AccessRequestReadySeconds.
				WithLabelValues("PodAccessRequest", "ns", "tmpl").
				Observe(12)

			Expect(testutil.CollectAndCount(AccessRequestReadySeconds)).To(BeNumerically(">=", 1))
		})
	})
})
