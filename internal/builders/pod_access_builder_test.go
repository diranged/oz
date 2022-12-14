package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/diranged/oz/internal/api/v1alpha1"
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
					APIReader: fakeClient,
					Request:   request,
					Template:  template,
				},
				Template: template,
				Request:  request,
			}
		})

		// TODO: Write tests that check the builder logic, more than that check the nested api logic
		It("generatePodTemplateSpec should return unmutated without error", func() {
			// Get the original pod template spec...
			podTemplateSpec, err := builder.generatePodTemplateSpec()
			Expect(err).To(Not(HaveOccurred()))

			// Run the PodSpec through the optional mutation config
			mutator := template.Spec.ControllerTargetMutationConfig
			ret, err := mutator.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// Wipe: metadata.labels (not optional)
			expectedPodTemplateSpec := podTemplateSpec.DeepCopy()
			expectedPodTemplateSpec.ObjectMeta.Labels = map[string]string{}

			// VERIFY: The original spec and new spec are identical
			Expect(ret.DeepCopy()).To(Equal(expectedPodTemplateSpec))
		})
	})
})
