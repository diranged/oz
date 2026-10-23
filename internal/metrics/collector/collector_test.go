package collector

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/metrics"
)

// erroringReader stands in for a client whose informer cache has not synced.
type erroringReader struct {
	client.Reader
}

func (erroringReader) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return errors.New("the cache is not started, can not read objects")
}

// series is one gathered sample, flattened into something convenient to assert
// against.
type series struct {
	Name   string
	Labels map[string]string
	Value  float64
}

// collect performs a scrape of a Collector backed by the supplied reader,
// against an isolated registry so that these assertions are unaffected by the
// process-wide controller-runtime registry.
func collect(reader client.Reader) []series {
	registry := prometheus.NewPedanticRegistry()
	ExpectWithOffset(1, registry.Register(&Collector{reader: reader})).To(Succeed())

	families, err := registry.Gather()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	gathered := []series{}
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			gathered = append(gathered, series{
				Name:   family.GetName(),
				Labels: labels,
				Value:  metric.GetGauge().GetValue(),
			})
		}
	}
	return gathered
}

// podRequest builds a PodAccessRequest as the collector would find it in the
// cluster - already stamped with a requester and with a settled readiness.
func podRequest(name, namespace, template, user string, ready bool) *v1alpha1.PodAccessRequest {
	req := &v1alpha1.PodAccessRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{v1alpha1.AnnotationRequestedBy: user},
		},
		Spec: v1alpha1.PodAccessRequestSpec{TemplateName: template},
	}
	req.Status.SetReady(ready)
	return req
}

func readerWith(objects ...client.Object) client.Reader {
	scheme := runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

var _ = Describe("Collector", func() {
	AfterEach(func() {
		metrics.SetIncludeUserLabel(false)
	})

	Context("with access requests in the cluster", func() {
		It("aggregates requests that share a label set", func() {
			metrics.SetIncludeUserLabel(true)

			gathered := collect(readerWith(
				podRequest("a", "prod", "debug-web", "dennis", true),
				podRequest("b", "prod", "debug-web", "dennis", true),
			))

			Expect(gathered).To(ConsistOf(series{
				Name: "oz_access_requests",
				Labels: map[string]string{
					"kind":      "PodAccessRequest",
					"namespace": "prod",
					"template":  "debug-web",
					"user":      "dennis",
					"ready":     "true",
				},
				Value: 2,
			}))
		})

		It("separates requests by template, user and readiness", func() {
			metrics.SetIncludeUserLabel(true)

			gathered := collect(readerWith(
				podRequest("a", "prod", "debug-web", "dennis", true),
				podRequest("b", "prod", "debug-web", "dennis", false),
				podRequest("c", "prod", "debug-api", "dennis", true),
				podRequest("d", "prod", "debug-web", "somebody-else", true),
			))

			Expect(gathered).To(HaveLen(4))
			for _, sample := range gathered {
				Expect(sample.Value).To(Equal(float64(1)))
			}
		})

		It("redacts the user label unless it has been enabled", func() {
			gathered := collect(readerWith(
				podRequest("a", "prod", "debug-web", "dennis", true),
			))

			Expect(gathered).To(HaveLen(1))
			Expect(gathered[0].Labels).To(HaveKeyWithValue("user", metrics.UserRedacted))
		})

		It("collapses distinct users into one series when redacting", func() {
			gathered := collect(readerWith(
				podRequest("a", "prod", "debug-web", "dennis", true),
				podRequest("b", "prod", "debug-web", "somebody-else", true),
			))

			Expect(gathered).To(HaveLen(1))
			Expect(gathered[0].Value).To(Equal(float64(2)))
		})

		It("labels each kind of request distinctly", func() {
			gathered := collect(readerWith(
				podRequest("a", "prod", "debug-web", "dennis", true),
				&v1alpha1.ExecAccessRequest{
					ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "prod"},
					Spec:       v1alpha1.ExecAccessRequestSpec{TemplateName: "exec-web"},
				},
			))

			kinds := []string{}
			for _, sample := range gathered {
				kinds = append(kinds, sample.Labels["kind"])
			}
			Expect(kinds).To(ConsistOf("PodAccessRequest", "ExecAccessRequest"))
		})

		It("reports requests that were never stamped with a requester", func() {
			metrics.SetIncludeUserLabel(true)

			gathered := collect(readerWith(&v1alpha1.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "prod"},
				Spec:       v1alpha1.PodAccessRequestSpec{TemplateName: "debug-web"},
			}))

			Expect(gathered).To(HaveLen(1))
			Expect(gathered[0].Labels).To(HaveKeyWithValue("user", metrics.UserUnknown))
		})
	})

	Context("with access templates in the cluster", func() {
		It("emits one series per template, named", func() {
			gathered := collect(readerWith(
				&v1alpha1.PodAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{Name: "debug-web", Namespace: "prod"},
				},
				&v1alpha1.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{Name: "exec-web", Namespace: "prod"},
				},
			))

			Expect(gathered).To(ConsistOf(
				series{
					Name: "oz_access_templates",
					Labels: map[string]string{
						"kind":      "PodAccessTemplate",
						"namespace": "prod",
						"name":      "debug-web",
						"ready":     "false",
					},
					Value: 1,
				},
				series{
					Name: "oz_access_templates",
					Labels: map[string]string{
						"kind":      "ExecAccessTemplate",
						"namespace": "prod",
						"name":      "exec-web",
						"ready":     "false",
					},
					Value: 1,
				},
			))
		})
	})

	Context("with an empty cluster", func() {
		It("reports no series rather than failing", func() {
			Expect(collect(readerWith())).To(BeEmpty())
		})
	})

	Context("when the cache cannot be read", func() {
		// A silent zero here would look identical to "nobody is using Oz",
		// which is exactly the wrong conclusion to let somebody draw.
		It("surfaces the failure as a scrape error", func() {
			registry := prometheus.NewPedanticRegistry()
			Expect(registry.Register(&Collector{reader: erroringReader{}})).To(Succeed())

			_, err := registry.Gather()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cache is not started"))
		})
	})
})
