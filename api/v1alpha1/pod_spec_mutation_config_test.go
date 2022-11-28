package v1alpha1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodSpecMutationConfig", Ordered, func() {
	ctx := context.Background()
	var podTemplateSpec corev1.PodTemplateSpec

	BeforeEach(func() {
		// Create a fake deployment target
		podTemplateSpec = v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"Foo": "bar",
				},
				Labels: map[string]string{
					"testLabel": "testValue",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "contA",
						Image: "nginx:latest",
					},
					{
						Name:  "contB",
						Image: "nginx:latest",
					},
				},
			},
		}
	})

	Context("Functions()", func() {
		It("getDefaultContainerID should return 0 by default", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{}

			// Run it
			ret, err := config.getDefaultContainerID(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: We got back '0' by default
			Expect(ret).To(Equal(0))
		})
		It(
			"getDefaultContainerID should return 1 by if we set the container name to contB",
			func() {
				// Basic resource with simple container identifier
				config := &PodTemplateSpecMutationConfig{DefaultContainerName: "contB"}

				// Run it
				ret, err := config.getDefaultContainerID(ctx, podTemplateSpec)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: We got back '1'
				Expect(ret).To(Equal(1))
			},
		)

		It(
			"getDefaultContainerID should return 1 by if the annotation is set",
			func() {
				// Patch the deployment
				podTemplateSpec.Annotations[DefaultContainerAnnotationKey] = "contB"

				// Basic resource with no mutation config
				config := &PodTemplateSpecMutationConfig{}

				// Run it
				ret, err := config.getDefaultContainerID(ctx, podTemplateSpec)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: We got back '1'
				Expect(ret).To(Equal(1))
			},
		)
	})
})
