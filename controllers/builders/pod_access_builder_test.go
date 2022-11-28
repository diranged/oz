package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/diranged/oz/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("PodAccessBuilder", Ordered, func() {
	Context("Functions()", func() {
		var (
			fakeClient client.Client
			deployment *appsv1.Deployment
			ctx        = context.Background()
			request    *api.PodAccessRequest
			template   *api.PodAccessTemplate
			builder    *PodAccessBuilder
		)

		BeforeEach(func() {
			// NOTE: Fake Client used here to make it easier to keep state separate between each It() test.
			fakeClient = fake.NewClientBuilder().WithRuntimeObjects().Build()

			// Create a fake deployment target
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dep",
					Namespace: "test-ns",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"testLabel": "testValue",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								api.DefaultContainerAnnotationKey: "contB",
								"Foo":                             "bar",
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
					},
				},
			}
			err := fakeClient.Create(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))

			// Create a default PodAccessTemplate. We'll mutate it for specific tests.
			template = &api.PodAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: deployment.Namespace,
				},
				Spec: api.PodAccessTemplateSpec{
					AccessConfig: api.AccessConfig{
						AllowedGroups:   []string{"testGroupA"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &api.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       deployment.Name,
					},
					ControllerTargetMutationConfig: &api.PodTemplateSpecMutationConfig{},
				},
			}
			err = fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create a simple PodAccessRequest resource to test the template with
			request = &api.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-req",
					Namespace: template.Namespace,
				},
				Spec: api.PodAccessRequestSpec{
					TemplateName: template.Name,
					Duration:     "5m",
				},
			}
			err = fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the PodAccessBuilder finally - fully populated with the
			// Request, Template and fake clients.
			builder = &PodAccessBuilder{
				BaseBuilder: BaseBuilder{
					Client:    fakeClient,
					Ctx:       ctx,
					Scheme:    &runtime.Scheme{},
					APIReader: fakeClient,
					Request:   request,
					Template:  template,
				},
				Template: template,
				Request:  request,
			}
		})

		It("generatePodTemplateSpec should return unmutated spec by default", func() {
			// Get the original pod template spec...
			podTemplateSpec, err := builder.generatePodTemplateSpec()
			Expect(err).To(Not(HaveOccurred()))

			// Run the PodSpec through the optional mutation config
			mutator := template.Spec.ControllerTargetMutationConfig
			mutatedSpec, err := mutator.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The original spec and new spec are identical
			Expect(podTemplateSpec).To(Equal(mutatedSpec))
		})

		It(
			"generatePodTemplateSpec should find the default container to mutate via annotation",
			func() {
				template.Spec.ControllerTargetMutationConfig = &api.PodTemplateSpecMutationConfig{
					Command: &[]string{"/bin/sleep"},
					Args:    &[]string{"100"},
				}
				// Get the original pod template spec...
				podTemplateSpec, err := builder.generatePodTemplateSpec()
				Expect(err).To(Not(HaveOccurred()))

				// Run the PodSpec through the optional mutation config
				mutator := template.Spec.ControllerTargetMutationConfig
				podTemplateSpec, err = mutator.PatchPodTemplateSpec(ctx, podTemplateSpec)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: contB (container 0) had some mutations
				Expect(len(podTemplateSpec.Spec.Containers[1].Command)).To(Equal(1))
				Expect(podTemplateSpec.Spec.Containers[1].Command[0]).To(Equal("/bin/sleep"))
				Expect(podTemplateSpec.Spec.Containers[1].Args[0]).To(Equal("100"))
			},
		)

		It(
			"generatePodTemplateSpec should find the default container based on the user supplied value",
			func() {
				// Create a default PodAccessTemplate. We'll mutate it for specific tests.
				cpuReq, _ := resource.ParseQuantity("1")
				template.Spec.ControllerTargetMutationConfig = &api.PodTemplateSpecMutationConfig{
					DefaultContainerName: "contA",
					Command:              &[]string{"/bin/sleep"},
					Args:                 &[]string{"100"},
					Env: []corev1.EnvVar{
						{Name: "FOO", Value: "BAR"},
					},
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							"cpu": cpuReq,
						},
					},
				}

				// Get the original pod template spec...
				podTemplateSpec, err := builder.generatePodTemplateSpec()
				Expect(err).To(Not(HaveOccurred()))

				// Run the PodSpec through the optional mutation config
				mutator := template.Spec.ControllerTargetMutationConfig
				podTemplateSpec, err = mutator.PatchPodTemplateSpec(ctx, podTemplateSpec)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: contA (container 0) had some mutations
				Expect(len(podTemplateSpec.Spec.Containers[0].Command)).To(Equal(1))
				Expect(podTemplateSpec.Spec.Containers[0].Command[0]).To(Equal("/bin/sleep"))
				Expect(podTemplateSpec.Spec.Containers[0].Args[0]).To(Equal("100"))
				Expect(podTemplateSpec.Spec.Containers[0].Env[0].Name).To(Equal("FOO"))
				Expect(podTemplateSpec.Spec.Containers[0].Env[0].Value).To(Equal("BAR"))
			},
		)
	})
})
