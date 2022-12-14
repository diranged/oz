package controllers

import (
	"context"
	"fmt"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers/builders"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type FakeBuilder struct {
	builders.BaseBuilder

	// Flags for faking GenerateAccessResources
	retStatusString string
	retAccessString string
	retErr          error
}

func (b *FakeBuilder) GenerateAccessResources() (statusString string, err error) {
	return b.retStatusString, b.retErr
}

var (
	_ builders.IBuilder = &FakeBuilder{}
	_ builders.IBuilder = (*FakeBuilder)(nil)
)

var _ = Describe("BaseReconciler", Ordered, func() {
	Context("Method Tests", func() {
		const TestName = "base-controller-test"

		var namespace *corev1.Namespace

		// Logger for our tests - makes it easier for us to debug sometimes
		ctx := context.Background()
		logger := log.FromContext(ctx)

		// These controller tests use a real Kubernetes backend and therefore they don't have a
		// significant amount of isolation between each test. We create one namespace at the
		// beginning of all of the tests for the duration of the tests.
		BeforeAll(func() {
			By("Creating the Namespace to perform the tests")
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: randomString(8),
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Refetch() should work", func() {
			// Initial test configmap
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TestName,
					Namespace: namespace.Name,
				},
				Data: map[string]string{
					"foo.data": "test data",
				},
			}

			// Base OzReconciler
			reconciler := &BaseReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				APIReader: k8sClient,
			}

			By("Creating a ConfigMap to reference for tests")
			// Create the config map and gather its resource version
			err := k8sClient.Create(ctx, cm)
			Expect(err).To(Not(HaveOccurred()))
			origResourceVer := cm.ResourceVersion
			logger.V(1).Info(fmt.Sprintf("Original ConfigMap ResourceVersion: %s", origResourceVer))

			// Now update the configmap...
			cm.Data = map[string]string{
				"foo.new": "test data",
			}
			err = k8sClient.Update(ctx, cm)
			Expect(err).To(Not(HaveOccurred()))
			newResourceVer := cm.ResourceVersion
			logger.V(1).Info(fmt.Sprintf("Updated ConfigMap ResourceVersion: %s", newResourceVer))

			// Verify that the two numbers do not match..
			Expect(origResourceVer).To(Not(Equal(newResourceVer)))

			// Now, refetch th data
			_, _ = reconciler.refetch(ctx, cm)

			// Verify that the new object has the new resource version
			Expect(newResourceVer).To(Equal(cm.ResourceVersion))
		})

		It("UpdateStatus() should work", func() {
			originalReq := &api.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TestName,
					Namespace: namespace.Name,
				},
				Spec: api.PodAccessRequestSpec{
					TemplateName: "Junk",
					Duration:     "1h",
				},
			}

			reconciler := &BaseReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				APIReader: k8sClient,
			}

			By("Creating an AccessRequest resource to update")
			err := k8sClient.Create(ctx, originalReq)
			Expect(err).To(Not(HaveOccurred()))

			By("Verifying the initial PodName is empty")
			Expect(originalReq.Status.PodName).To(Equal(""))

			By("Set the Status.PodName to something")
			originalReq.Status.PodName = "bogus"
			err = reconciler.updateStatus(ctx, originalReq)
			Expect(err).To(Not(HaveOccurred()))

			By("Get a new reference to the AccessRequest, verify the PodName status")
			freshReq := &api.PodAccessRequest{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      TestName,
				Namespace: namespace.Name,
			}, freshReq)
			Expect(err).To(Not(HaveOccurred()))
			Expect(freshReq.Status.PodName).To(Equal("bogus"))
		})

		It("UpdateStatus() should return failures properly", func() {
			request := &api.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-fail-test", TestName),
					Namespace: namespace.Name,
				},
				Spec: api.PodAccessRequestSpec{
					TemplateName: "Junk",
					Duration:     "1h",
				},
			}

			reconciler := &BaseReconciler{
				Client:    client.NewNamespacedClient(k8sClient, "bogus"),
				Scheme:    k8sClient.Scheme(),
				APIReader: client.NewNamespacedClient(k8sClient, "bogus"),
			}

			By("Creating an AccessRequest resource to update")
			err := k8sClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			By("Now try to update the request with a narrowly scoped namespaced client")
			err = reconciler.updateStatus(ctx, request)
			Expect(err).To(HaveOccurred())
		})
	})
})
