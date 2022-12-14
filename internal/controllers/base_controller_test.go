package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/diranged/oz/internal/builders"
	testingutils "github.com/diranged/oz/internal/testing/utils"
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
		// logger := log.FromContext(ctx)

		// These controller tests use a real Kubernetes backend and therefore they don't have a
		// significant amount of isolation between each test. We create one namespace at the
		// beginning of all of the tests for the duration of the tests.
		BeforeAll(func() {
			By("Creating the Namespace to perform the tests")
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testingutils.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
