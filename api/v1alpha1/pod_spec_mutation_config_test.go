package v1alpha1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

		It("PatchPodTemplateSpec should return unmutated by default", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: Unmutated by default
			Expect(ret).To(Equal(podTemplateSpec))
		})

		It("PatchPodTemplateSpec should mutate command", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{
				Command: &[]string{"/bin/sleep"},
				Args:    &[]string{"100"},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewQuantity(1, resource.Format("DecimalExponent")),
					},
				},
				Env: []corev1.EnvVar{
					{Name: "FOO", Value: "BAR"},
				},
			}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: Command/Args is set
			Expect(ret.Spec.Containers[0].Command[0]).To(Equal("/bin/sleep"))
			Expect(ret.Spec.Containers[0].Args[0]).To(Equal("100"))

			// VERIFY: Resources are set
			Expect(ret.Spec.Containers[0].Resources.Limits.Cpu().String()).To(Equal("1"))

			// VERIFY: EnvVar is set
			Expect(len(ret.Spec.Containers[0].Env)).To(Equal(1))
		})

		It("PatchPodTemplateSpec should fail if invalid container name supplied", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{DefaultContainerName: "bogus"}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(HaveOccurred())

			// VERIFY: Unmutated
			Expect(ret).To(Equal(podTemplateSpec))
		})
	})
})
